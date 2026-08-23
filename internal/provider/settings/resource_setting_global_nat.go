package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// globalNatModel represents the global NAT (Network Address Translation) settings
// for a UniFi site.
type globalNatModel struct {
	base.Model
	ExcludedNetworkIDs types.List   `tfsdk:"excluded_network_ids"`
	Mode               types.String `tfsdk:"mode"`
}

func (d *globalNatModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingGlobalNat{
		ID: d.ID.ValueString(),
	}

	if !ut.IsEmptyString(d.Mode) {
		model.Mode = d.Mode.ValueString()
	}

	if !d.ExcludedNetworkIDs.IsNull() {
		var excluded []string
		diags.Append(ut.ListElementsAs(ctx, d.ExcludedNetworkIDs, &excluded)...)
		if diags.HasError() {
			return nil, diags
		}
		model.ExcludedNetworkIDs = excluded
	}

	return model, diags
}

func (d *globalNatModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingGlobalNat)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingGlobalNat")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Mode = ut.StringOrNull(model.Mode)

	if len(model.ExcludedNetworkIDs) > 0 {
		list, ld := types.ListValueFrom(ctx, types.StringType, model.ExcludedNetworkIDs)
		diags.Append(ld...)
		if diags.HasError() {
			return diags
		}
		d.ExcludedNetworkIDs = list
	} else {
		d.ExcludedNetworkIDs = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel               = &globalNatModel{}
	_ resource.Resource                = &globalNatResource{}
	_ resource.ResourceWithConfigure   = &globalNatResource{}
	_ resource.ResourceWithImportState = &globalNatResource{}
)

type globalNatResource struct {
	*base.GenericResource[*globalNatModel]
}

func (r *globalNatResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_global_nat` resource manages the global NAT (Network Address Translation) " +
			"settings for a UniFi site. It controls the NAT mode applied site-wide and lets you exclude specific networks " +
			"from global NAT.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"excluded_network_ids": schema.ListAttribute{
				MarkdownDescription: "List of network IDs that are excluded from global NAT.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "The global NAT mode. Valid values are:\n" +
					"* `auto` - Automatically manage NAT for site networks.\n" +
					"* `custom` - Use a custom NAT configuration.\n" +
					"* `off` - Disable global NAT.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "custom", "off"),
				},
			},
		},
	}
}

// NewGlobalNatResource creates a new instance of the global NAT setting resource.
func NewGlobalNatResource() resource.Resource {
	r := &globalNatResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_global_nat",
		func() *globalNatModel { return &globalNatModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingGlobalNat(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingGlobalNat)
			return client.UpdateSettingGlobalNat(ctx, site, b)
		},
	)
	return r
}
