package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// elementAdoptModel represents the Element Adoption settings for a UniFi site,
// which allow UniFi Elements devices to be adopted using a dedicated SSID and
// pre-shared key.
type elementAdoptModel struct {
	base.Model
	Enabled       types.Bool   `tfsdk:"enabled"`
	XElementEssid types.String `tfsdk:"x_element_essid"`
	XElementPsk   types.String `tfsdk:"x_element_psk"`
}

func (d *elementAdoptModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingElementAdopt{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// Only set optional fields if element adoption is enabled
	if d.Enabled.ValueBool() {
		if !ut.IsEmptyString(d.XElementEssid) {
			model.XElementEssid = d.XElementEssid.ValueString()
		}
		if !ut.IsEmptyString(d.XElementPsk) {
			model.XElementPsk = d.XElementPsk.ValueString()
		}
	}

	return model, diags
}

func (d *elementAdoptModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingElementAdopt)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingElementAdopt")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if element adoption is enabled
	if model.Enabled {
		d.XElementEssid = ut.StringOrNull(model.XElementEssid)
		d.XElementPsk = ut.StringOrNull(model.XElementPsk)
	} else {
		d.XElementEssid = types.StringNull()
		d.XElementPsk = types.StringNull()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &elementAdoptModel{}
	_ resource.Resource                     = &elementAdoptResource{}
	_ resource.ResourceWithConfigure        = &elementAdoptResource{}
	_ resource.ResourceWithImportState      = &elementAdoptResource{}
	_ resource.ResourceWithConfigValidators = &elementAdoptResource{}
)

type elementAdoptResource struct {
	*base.GenericResource[*elementAdoptModel]
}

func (r *elementAdoptResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false), path.MatchRoot("x_element_essid"), path.MatchRoot("x_element_psk")),
	}
}

func (r *elementAdoptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_element_adopt` resource manages Element Adoption settings for a UniFi site, allowing UniFi " +
			"Elements devices to be adopted over a dedicated SSID and pre-shared key.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Element adoption is enabled.",
				Required:            true,
			},
			"x_element_essid": schema.StringAttribute{
				MarkdownDescription: "The SSID used for Element adoption. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"x_element_psk": schema.StringAttribute{
				MarkdownDescription: "The pre-shared key used for Element adoption. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// NewElementAdoptResource creates a new instance of the Element Adoption setting resource.
func NewElementAdoptResource() resource.Resource {
	r := &elementAdoptResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_element_adopt",
		func() *elementAdoptModel { return &elementAdoptModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingElementAdopt(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingElementAdopt)
			return client.UpdateSettingElementAdopt(ctx, site, b)
		},
	)
	return r
}
