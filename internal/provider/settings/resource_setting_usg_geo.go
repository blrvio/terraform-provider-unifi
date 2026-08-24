package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// usgGeoIPFilteringModel is the nested Region Blocking (Geo IP filtering)
// configuration stored under the `usg_geo` setting's `ip_filtering` object.
type usgGeoIPFilteringModel struct {
	Action           types.String `tfsdk:"action"`
	Countries        types.List   `tfsdk:"countries"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	TrafficDirection types.String `tfsdk:"traffic_direction"`
}

func (m *usgGeoIPFilteringModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"action":            types.StringType,
		"countries":         types.ListType{ElemType: types.StringType},
		"enabled":           types.BoolType,
		"traffic_direction": types.StringType,
	}
}

// usgGeoModel represents the `usg_geo` setting (Region Blocking) for a UniFi
// site. On Network 10.x this lives in a setting key SEPARATE from `usg`, with a
// nested `ip_filtering` object — the flat geo fields on `unifi_setting_usg` do
// not exist on the current controller.
type usgGeoModel struct {
	base.Model
	IPFiltering types.Object `tfsdk:"ip_filtering"`
}

func (d *usgGeoModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingUsgGeo{
		ID: d.ID.ValueString(),
	}

	if ut.IsDefined(d.IPFiltering) {
		var ipf usgGeoIPFilteringModel
		diags.Append(d.IPFiltering.As(ctx, &ipf, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		countries, cDiags := ut.ListElementsToString(ctx, ipf.Countries)
		diags.Append(cDiags...)
		if diags.HasError() {
			return nil, diags
		}

		model.IPFiltering = unifi.SettingUsgGeoIPFiltering{
			Action:           ipf.Action.ValueString(),
			Countries:        countries,
			Enabled:          ipf.Enabled.ValueBool(),
			TrafficDirection: ipf.TrafficDirection.ValueString(),
		}
	}

	return model, diags
}

func (d *usgGeoModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingUsgGeo)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingUsgGeo")
		return diags
	}

	d.ID = types.StringValue(model.ID)

	countries, cDiags := ut.StringToListElements(ctx, model.IPFiltering.Countries)
	diags.Append(cDiags...)
	if diags.HasError() {
		return diags
	}

	ipf := &usgGeoIPFilteringModel{
		Action:           types.StringValue(model.IPFiltering.Action),
		Countries:        countries,
		Enabled:          types.BoolValue(model.IPFiltering.Enabled),
		TrafficDirection: types.StringValue(model.IPFiltering.TrafficDirection),
	}

	obj, oDiags := types.ObjectValueFrom(ctx, ipf.AttributeTypes(), ipf)
	diags.Append(oDiags...)
	if diags.HasError() {
		return diags
	}
	d.IPFiltering = obj

	return diags
}

var (
	_ base.ResourceModel               = &usgGeoModel{}
	_ resource.Resource                = &usgGeoResource{}
	_ resource.ResourceWithConfigure   = &usgGeoResource{}
	_ resource.ResourceWithImportState = &usgGeoResource{}
)

type usgGeoResource struct {
	*base.GenericResource[*usgGeoModel]
}

func (r *usgGeoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_usg_geo` resource manages Region Blocking (Geo IP filtering) for a UniFi " +
			"gateway. On UniFi Network 10.x this is stored in a dedicated `usg_geo` setting (separate from " +
			"`unifi_setting_usg`), with a nested `ip_filtering` object. Use it to block or allow traffic by country of " +
			"origin.\n\n" +
			"This resource replaces the deprecated `geo_ip_filtering` block on `unifi_setting_usg`, which mapped to " +
			"fields that no longer exist on current controllers.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"ip_filtering": schema.SingleNestedAttribute{
				MarkdownDescription: "Region Blocking configuration. Uses IP geolocation to block or allow traffic based on " +
					"the country of origin.",
				Required: true,
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						MarkdownDescription: "Whether the listed `countries` are blocked or allowed. Valid values are " +
							"`block` (default — block the listed countries, allow all others) or `allow` (allow only the " +
							"listed countries, block all others).",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("block"),
						Validators: []validator.String{
							stringvalidator.OneOf("block", "allow"),
						},
					},
					"countries": schema.ListAttribute{
						MarkdownDescription: "List of two-letter ISO 3166-1 alpha-2 country codes to block or allow " +
							"(e.g. `[\"RU\", \"CN\"]`). Must contain at least one country code.",
						Required:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(validators.CountryCodeAlpha2()),
						},
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether Region Blocking is enabled. Defaults to `true`.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"traffic_direction": schema.StringAttribute{
						MarkdownDescription: "Which traffic direction the filter applies to. Valid values are `both` " +
							"(default), `ingress` (incoming only), or `egress` (outgoing only).",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("both"),
						Validators: []validator.String{
							stringvalidator.OneOf("both", "ingress", "egress"),
						},
					},
				},
			},
		},
	}
}

// NewUsgGeoResource creates a new instance of the USG Region Blocking setting resource.
func NewUsgGeoResource() resource.Resource {
	r := &usgGeoResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_usg_geo",
		func() *usgGeoModel { return &usgGeoModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingUsgGeo(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingUsgGeo)
			return client.UpdateSettingUsgGeo(ctx, site, b)
		},
	)
	return r
}
