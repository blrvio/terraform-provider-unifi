package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// baresipModel represents the Baresip (SIP client) settings for a UniFi site.
type baresipModel struct {
	base.Model
	Enabled       types.Bool   `tfsdk:"enabled"`
	OutboundProxy types.String `tfsdk:"outbound_proxy"`
	PackageURL    types.String `tfsdk:"package_url"`
	Server        types.String `tfsdk:"server"`
}

func (d *baresipModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingBaresip{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// Only set optional fields if baresip is enabled
	if d.Enabled.ValueBool() {
		if !ut.IsEmptyString(d.OutboundProxy) {
			model.OutboundProxy = d.OutboundProxy.ValueString()
		}
		if !ut.IsEmptyString(d.PackageURL) {
			model.PackageUrl = d.PackageURL.ValueString()
		}
		if !ut.IsEmptyString(d.Server) {
			model.Server = d.Server.ValueString()
		}
	}

	return model, diags
}

func (d *baresipModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingBaresip)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingBaresip")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if baresip is enabled
	if model.Enabled {
		d.OutboundProxy = ut.StringOrNull(model.OutboundProxy)
		d.PackageURL = ut.StringOrNull(model.PackageUrl)
		d.Server = ut.StringOrNull(model.Server)
	} else {
		d.OutboundProxy = types.StringNull()
		d.PackageURL = types.StringNull()
		d.Server = types.StringNull()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &baresipModel{}
	_ resource.Resource                     = &baresipResource{}
	_ resource.ResourceWithConfigure        = &baresipResource{}
	_ resource.ResourceWithImportState      = &baresipResource{}
	_ resource.ResourceWithConfigValidators = &baresipResource{}
)

type baresipResource struct {
	*base.GenericResource[*baresipModel]
}

func (r *baresipResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false), path.MatchRoot("outbound_proxy"), path.MatchRoot("package_url"), path.MatchRoot("server")),
	}
}

func (r *baresipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_baresip` resource manages the Baresip SIP client settings for a UniFi site.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Baresip SIP client is enabled.",
				Required:            true,
			},
			"outbound_proxy": schema.StringAttribute{
				MarkdownDescription: "The outbound SIP proxy to route calls through. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"package_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the Baresip package to install. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.URL(),
				},
			},
			"server": schema.StringAttribute{
				MarkdownDescription: "The SIP server to register with. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewBaresipResource creates a new instance of the Baresip setting resource.
func NewBaresipResource() resource.Resource {
	r := &baresipResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_baresip",
		func() *baresipModel { return &baresipModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingBaresip(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingBaresip)
			return client.UpdateSettingBaresip(ctx, site, b)
		},
	)
	return r
}
