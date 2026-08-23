package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// portaModel represents the Porta settings for a UniFi site, controlling the
// secondary WAN (WAN2) port on the UniFi Security Gateway 3P (USG-3P).
type portaModel struct {
	base.Model
	Ugw3WAN2Enabled types.Bool `tfsdk:"ugw3_wan2_enabled"`
}

func (d *portaModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingPorta{
		ID:              d.ID.ValueString(),
		Ugw3WAN2Enabled: d.Ugw3WAN2Enabled.ValueBool(),
	}

	return model, diags
}

func (d *portaModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingPorta)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingPorta")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Ugw3WAN2Enabled = types.BoolValue(model.Ugw3WAN2Enabled)

	return diags
}

var (
	_ base.ResourceModel               = &portaModel{}
	_ resource.Resource                = &portaResource{}
	_ resource.ResourceWithConfigure   = &portaResource{}
	_ resource.ResourceWithImportState = &portaResource{}
)

type portaResource struct {
	*base.GenericResource[*portaModel]
}

func (r *portaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_porta` resource manages Porta settings for a UniFi site, controlling the secondary WAN " +
			"(WAN2) port on the UniFi Security Gateway 3P (USG-3P).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"ugw3_wan2_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the WAN2 port of the UniFi Security Gateway 3P (USG-3P) is enabled.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewPortaResource creates a new instance of the Porta setting resource.
func NewPortaResource() resource.Resource {
	r := &portaResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_porta",
		func() *portaModel { return &portaModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingPorta(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingPorta)
			return client.UpdateSettingPorta(ctx, site, b)
		},
	)
	return r
}
