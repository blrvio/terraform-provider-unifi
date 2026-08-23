package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// superIdentityModel represents the identity settings (console hostname and
// display name) of a UniFi site's controller.
type superIdentityModel struct {
	base.Model
	Hostname types.String `tfsdk:"hostname"`
	Name     types.String `tfsdk:"name"`
}

func (d *superIdentityModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperIdentity{
		ID: d.ID.ValueString(),
	}
	if !ut.IsEmptyString(d.Hostname) {
		model.Hostname = d.Hostname.ValueString()
	}
	if !ut.IsEmptyString(d.Name) {
		model.Name = d.Name.ValueString()
	}

	return model, diags
}

func (d *superIdentityModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperIdentity)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperIdentity")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Hostname = ut.StringOrNull(model.Hostname)
	d.Name = ut.StringOrNull(model.Name)

	return diags
}

var (
	_ base.ResourceModel               = &superIdentityModel{}
	_ resource.Resource                = &superIdentityResource{}
	_ resource.ResourceWithConfigure   = &superIdentityResource{}
	_ resource.ResourceWithImportState = &superIdentityResource{}
)

type superIdentityResource struct {
	*base.GenericResource[*superIdentityModel]
}

func (r *superIdentityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_identity` resource manages the identity settings of a UniFi console, " +
			"such as its network hostname and human-readable display name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The network hostname of the UniFi console. Must be a valid hostname.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.Hostname(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The human-readable display name of the UniFi console.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewSuperIdentityResource creates a new instance of the super identity setting resource.
func NewSuperIdentityResource() resource.Resource {
	r := &superIdentityResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_identity",
		func() *superIdentityModel { return &superIdentityModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperIdentity(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperIdentity)
			return client.UpdateSettingSuperIdentity(ctx, site, b)
		},
	)
	return r
}
