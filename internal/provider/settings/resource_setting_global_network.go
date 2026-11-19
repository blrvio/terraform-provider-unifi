package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

type globalNetworkModel struct {
	base.Model
	DefaultSecurityPosture types.String `tfsdk:"default_security_posture"`
}

func (d *globalNetworkModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.SettingGlobalNetwork{
		ID:                     d.ID.ValueString(),
		DefaultSecurityPosture: d.DefaultSecurityPosture.ValueString(),
	}, nil
}

func (d *globalNetworkModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	model, ok := other.(*unifi.SettingGlobalNetwork)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingGlobalNetwork")
		return diags
	}
	d.ID = types.StringValue(model.ID)
	d.DefaultSecurityPosture = types.StringValue(model.DefaultSecurityPosture)
	return diags
}

var (
	_ base.ResourceModel               = &globalNetworkModel{}
	_ resource.Resource                = &globalNetworkResource{}
	_ resource.ResourceWithConfigure   = &globalNetworkResource{}
	_ resource.ResourceWithImportState = &globalNetworkResource{}
)

type globalNetworkResource struct {
	*base.GenericResource[*globalNetworkModel]
}

func (r *globalNetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site-wide `global_network` setting (UniFi Network 11.x), which holds the default " +
			"security posture of the zone-based firewall.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"default_security_posture": schema.StringAttribute{
				MarkdownDescription: "The default security posture between zones, as stored by the controller (`ALLOW_ALL` on a " +
					"default UniFi Network 11.0.99 site).",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
		},
	}
}

// NewGlobalNetworkResource creates the unifi_setting_global_network resource.
func NewGlobalNetworkResource() resource.Resource {
	r := &globalNetworkResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_global_network",
		unifi.SettingGlobalNetworkKey,
		func() *globalNetworkModel { return &globalNetworkModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingGlobalNetwork(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingGlobalNetwork)
			return client.UpdateSettingGlobalNetwork(ctx, site, b)
		},
	)
	return r
}
