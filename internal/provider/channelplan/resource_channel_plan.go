package channelplan

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// radioTableItemModel is one per-radio entry in a channel plan.
type radioTableItemModel struct {
	Channel     types.String `tfsdk:"channel"`
	DeviceMAC   types.String `tfsdk:"device_mac"`
	Name        types.String `tfsdk:"name"`
	TxPower     types.String `tfsdk:"tx_power"`
	TxPowerMode types.String `tfsdk:"tx_power_mode"`
	Width       types.Int32  `tfsdk:"width"`
}

func (m radioTableItemModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"channel":       types.StringType,
		"device_mac":    types.StringType,
		"name":          types.StringType,
		"tx_power":      types.StringType,
		"tx_power_mode": types.StringType,
		"width":         types.Int32Type,
	}
}

func radioTableObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: radioTableItemModel{}.AttributeTypes()}
}

// ChannelPlanModel represents the data model for a UniFi channel plan.
type ChannelPlanModel struct {
	base.Model
	Date       types.String `tfsdk:"date"`
	RadioTable types.List   `tfsdk:"radio_table"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *ChannelPlanModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var items []radioTableItemModel
	if ut.IsDefined(m.RadioTable) {
		diags.Append(ut.ListElementsAs(ctx, m.RadioTable, &items)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	radios := make([]unifi.ChannelPlanRadioTable, 0, len(items))
	for _, it := range items {
		radios = append(radios, unifi.ChannelPlanRadioTable{
			Channel:     it.Channel.ValueString(),
			DeviceMAC:   utils.CleanMAC(it.DeviceMAC.ValueString()),
			Name:        it.Name.ValueString(),
			TxPower:     it.TxPower.ValueString(),
			TxPowerMode: it.TxPowerMode.ValueString(),
			Width:       int(it.Width.ValueInt32()),
		})
	}

	return &unifi.ChannelPlan{
		ID:         m.ID.ValueString(),
		Date:       m.Date.ValueString(),
		RadioTable: radios,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *ChannelPlanModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.ChannelPlan)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.ChannelPlan, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Date = types.StringValue(model.Date)

	items := make([]radioTableItemModel, 0, len(model.RadioTable))
	for _, r := range model.RadioTable {
		items = append(items, radioTableItemModel{
			Channel:     ut.StringOrNull(r.Channel),
			DeviceMAC:   ut.StringOrNull(r.DeviceMAC),
			Name:        ut.StringOrNull(r.Name),
			TxPower:     ut.StringOrNull(r.TxPower),
			TxPowerMode: ut.StringOrNull(r.TxPowerMode),
			Width:       ut.Int32OrNull(r.Width),
		})
	}
	list, diags := types.ListValueFrom(ctx, radioTableObjectType(), items)
	m.RadioTable = list
	return diags
}

var (
	_ resource.Resource                = &channelPlanResource{}
	_ resource.ResourceWithConfigure   = &channelPlanResource{}
	_ resource.ResourceWithImportState = &channelPlanResource{}
	_ base.Resource                    = &channelPlanResource{}
	_ base.ResourceModel               = &ChannelPlanModel{}
)

type channelPlanResource struct {
	*base.GenericResource[*ChannelPlanModel]
}

// NewChannelPlanResource creates a new instance of the channel plan resource.
func NewChannelPlanResource() resource.Resource {
	return &channelPlanResource{
		GenericResource: base.NewGenericResource(
			"unifi_channel_plan",
			func() *ChannelPlanModel { return &ChannelPlanModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetChannelPlan(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					plan, _ := model.(*unifi.ChannelPlan)
					return client.CreateChannelPlan(ctx, site, plan)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					plan, _ := model.(*unifi.ChannelPlan)
					return client.UpdateChannelPlan(ctx, site, plan)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteChannelPlan(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *channelPlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_channel_plan` resource manages a channel plan in the UniFi controller.\n\n" +
			"~> **Advanced.** A channel plan is normally the *output* of the controller's RF channel " +
			"optimization: `date` is the generation timestamp and `radio_table` assigns a channel, tx-power " +
			"and width to each radio, keyed by the runtime `device_mac`. Manage this resource only when you " +
			"intend to author a fixed channel plan explicitly; otherwise let the controller generate it.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"date": schema.StringAttribute{
				MarkdownDescription: "The channel plan generation timestamp (RFC3339-like `YYYY-MM-DDTHH:MM:SSZ`), " +
					"or empty. Usually set by the controller when the plan is generated.",
				Optional: true,
				Computed: true,
			},
			"radio_table": schema.ListNestedAttribute{
				MarkdownDescription: "The per-radio channel assignments that make up the plan.",
				Optional:            true,
				Computed:            true,
				Default:             ut.DefaultEmptyList(radioTableObjectType()),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"device_mac": schema.StringAttribute{
							MarkdownDescription: "The MAC address of the radio's device.",
							Optional:            true,
							Validators: []validator.String{
								validators.Mac,
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The radio name (for example `ra0`, `rai0`).",
							Optional:            true,
						},
						"channel": schema.StringAttribute{
							MarkdownDescription: "The assigned channel, or `auto`.",
							Optional:            true,
						},
						"tx_power": schema.StringAttribute{
							MarkdownDescription: "The transmit power (a number) or `auto`.",
							Optional:            true,
						},
						"tx_power_mode": schema.StringAttribute{
							MarkdownDescription: "The transmit power mode. One of `auto`, `medium`, `high`, `low`, `custom`.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("auto", "medium", "high", "low", "custom"),
							},
						},
						"width": schema.Int32Attribute{
							MarkdownDescription: "The channel width in MHz. One of `20`, `40`, `80`, `160`.",
							Optional:            true,
							Validators: []validator.Int32{
								int32validator.OneOf(20, 40, 80, 160),
							},
						},
					},
				},
			},
		},
	}
}
