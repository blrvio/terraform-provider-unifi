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

// superMailModel represents the outgoing mail provider selection used by the
// UniFi console for sending notifications.
type superMailModel struct {
	base.Model
	Provider types.String `tfsdk:"provider_type"`
}

func (d *superMailModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperMail{
		ID: d.ID.ValueString(),
	}
	if !ut.IsEmptyString(d.Provider) {
		model.Provider = d.Provider.ValueString()
	}

	return model, diags
}

func (d *superMailModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperMail)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperMail")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Provider = ut.StringOrNull(model.Provider)

	return diags
}

var (
	_ base.ResourceModel               = &superMailModel{}
	_ resource.Resource                = &superMailResource{}
	_ resource.ResourceWithConfigure   = &superMailResource{}
	_ resource.ResourceWithImportState = &superMailResource{}
)

type superMailResource struct {
	*base.GenericResource[*superMailModel]
}

func (r *superMailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_mail` resource selects the outgoing mail provider used by the UniFi " +
			"console to deliver email notifications.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "The outgoing mail provider. Valid values are:\n" +
					"* `smtp` - Use a custom SMTP server (configured via `unifi_setting_super_smtp`)\n" +
					"* `cloud` - Use Ubiquiti's cloud mail relay\n" +
					"* `disabled` - Do not send email",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("smtp", "cloud", "disabled"),
				},
			},
		},
	}
}

// NewSuperMailResource creates a new instance of the super mail setting resource.
func NewSuperMailResource() resource.Resource {
	r := &superMailResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_mail",
		func() *superMailModel { return &superMailModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperMail(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperMail)
			return client.UpdateSettingSuperMail(ctx, site, b)
		},
	)
	return r
}
