package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// globalApModel represents the global access point (radio) settings for a UniFi
// site, controlling default channel width and transmit power for each radio band.
type globalApModel struct {
	base.Model
	ApExclusions    types.List   `tfsdk:"ap_exclusions"`
	NaChannelSize   types.Int32  `tfsdk:"na_channel_size"`
	NaTxPower       types.Int32  `tfsdk:"na_tx_power"`
	NaTxPowerMode   types.String `tfsdk:"na_tx_power_mode"`
	NgChannelSize   types.Int32  `tfsdk:"ng_channel_size"`
	NgTxPower       types.Int32  `tfsdk:"ng_tx_power"`
	NgTxPowerMode   types.String `tfsdk:"ng_tx_power_mode"`
	SixEChannelSize types.Int32  `tfsdk:"six_e_channel_size"`
	SixETxPower     types.Int32  `tfsdk:"six_e_tx_power"`
	SixETxPowerMode types.String `tfsdk:"six_e_tx_power_mode"`
}

func (d *globalApModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingGlobalAp{
		ID: d.ID.ValueString(),
	}

	if !d.ApExclusions.IsNull() {
		var exclusions []string
		diags.Append(ut.ListElementsAs(ctx, d.ApExclusions, &exclusions)...)
		if diags.HasError() {
			return nil, diags
		}
		model.ApExclusions = exclusions
	}

	if !d.NaChannelSize.IsNull() {
		model.NaChannelSize = int(d.NaChannelSize.ValueInt32())
	}
	if !d.NaTxPower.IsNull() {
		model.NaTxPower = int(d.NaTxPower.ValueInt32())
	}
	if !ut.IsEmptyString(d.NaTxPowerMode) {
		model.NaTxPowerMode = d.NaTxPowerMode.ValueString()
	}
	if !d.NgChannelSize.IsNull() {
		model.NgChannelSize = int(d.NgChannelSize.ValueInt32())
	}
	if !d.NgTxPower.IsNull() {
		model.NgTxPower = int(d.NgTxPower.ValueInt32())
	}
	if !ut.IsEmptyString(d.NgTxPowerMode) {
		model.NgTxPowerMode = d.NgTxPowerMode.ValueString()
	}
	if !d.SixEChannelSize.IsNull() {
		model.SixEChannelSize = int(d.SixEChannelSize.ValueInt32())
	}
	if !d.SixETxPower.IsNull() {
		model.SixETxPower = int(d.SixETxPower.ValueInt32())
	}
	if !ut.IsEmptyString(d.SixETxPowerMode) {
		model.SixETxPowerMode = d.SixETxPowerMode.ValueString()
	}

	return model, diags
}

func (d *globalApModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingGlobalAp)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingGlobalAp")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.NaChannelSize = ut.Int32OrNull(model.NaChannelSize)
	d.NaTxPower = ut.Int32OrNull(model.NaTxPower)
	d.NaTxPowerMode = ut.StringOrNull(model.NaTxPowerMode)
	d.NgChannelSize = ut.Int32OrNull(model.NgChannelSize)
	d.NgTxPower = ut.Int32OrNull(model.NgTxPower)
	d.NgTxPowerMode = ut.StringOrNull(model.NgTxPowerMode)
	d.SixEChannelSize = ut.Int32OrNull(model.SixEChannelSize)
	d.SixETxPower = ut.Int32OrNull(model.SixETxPower)
	d.SixETxPowerMode = ut.StringOrNull(model.SixETxPowerMode)

	if len(model.ApExclusions) > 0 {
		list, ld := types.ListValueFrom(ctx, types.StringType, model.ApExclusions)
		diags.Append(ld...)
		if diags.HasError() {
			return diags
		}
		d.ApExclusions = list
	} else {
		d.ApExclusions = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel               = &globalApModel{}
	_ resource.Resource                = &globalApResource{}
	_ resource.ResourceWithConfigure   = &globalApResource{}
	_ resource.ResourceWithImportState = &globalApResource{}
)

type globalApResource struct {
	*base.GenericResource[*globalApModel]
}

func (r *globalApResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	txPowerModeValidators := func() []validator.String {
		return []validator.String{
			stringvalidator.OneOf("auto", "medium", "high", "low", "custom"),
		}
	}
	txPowerValidators := func() []validator.Int32 {
		return []validator.Int32{
			int32validator.Between(0, 49),
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_global_ap` resource manages site-wide default access point radio settings, " +
			"including channel width and transmit power for the 5 GHz (`na`), 2.4 GHz (`ng`), and 6 GHz (`6e`) bands.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"ap_exclusions": schema.ListAttribute{
				MarkdownDescription: "List of access point MAC addresses excluded from these global radio settings.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"na_channel_size": schema.Int32Attribute{
				MarkdownDescription: "Default channel width (MHz) for the 5 GHz (`na`) band. Valid values: 20, 40, 80, 160.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.OneOf(20, 40, 80, 160),
				},
			},
			"na_tx_power": schema.Int32Attribute{
				MarkdownDescription: "Default transmit power (dBm) for the 5 GHz (`na`) band. Valid values: 0-49.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerValidators(),
			},
			"na_tx_power_mode": schema.StringAttribute{
				MarkdownDescription: "Transmit power mode for the 5 GHz (`na`) band. Valid values: `auto`, `medium`, `high`, `low`, `custom`.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerModeValidators(),
			},
			"ng_channel_size": schema.Int32Attribute{
				MarkdownDescription: "Default channel width (MHz) for the 2.4 GHz (`ng`) band. Valid values: 20, 40.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.OneOf(20, 40),
				},
			},
			"ng_tx_power": schema.Int32Attribute{
				MarkdownDescription: "Default transmit power (dBm) for the 2.4 GHz (`ng`) band. Valid values: 0-49.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerValidators(),
			},
			"ng_tx_power_mode": schema.StringAttribute{
				MarkdownDescription: "Transmit power mode for the 2.4 GHz (`ng`) band. Valid values: `auto`, `medium`, `high`, `low`, `custom`.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerModeValidators(),
			},
			"six_e_channel_size": schema.Int32Attribute{
				MarkdownDescription: "Default channel width (MHz) for the 6 GHz (`6e`) band. Valid values: 20, 40, 80, 160.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.OneOf(20, 40, 80, 160),
				},
			},
			"six_e_tx_power": schema.Int32Attribute{
				MarkdownDescription: "Default transmit power (dBm) for the 6 GHz (`6e`) band. Valid values: 0-49.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerValidators(),
			},
			"six_e_tx_power_mode": schema.StringAttribute{
				MarkdownDescription: "Transmit power mode for the 6 GHz (`6e`) band. Valid values: `auto`, `medium`, `high`, `low`, `custom`.",
				Optional:            true,
				Computed:            true,
				Validators:          txPowerModeValidators(),
			},
		},
	}
}

// NewGlobalApResource creates a new instance of the global AP setting resource.
func NewGlobalApResource() resource.Resource {
	r := &globalApResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_global_ap",
		func() *globalApModel { return &globalApModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingGlobalAp(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingGlobalAp)
			return client.UpdateSettingGlobalAp(ctx, site, b)
		},
	)
	return r
}
