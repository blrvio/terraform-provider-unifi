package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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

// superSMTPModel represents the custom SMTP server settings used by the UniFi
// console to deliver email notifications.
type superSMTPModel struct {
	base.Model
	Enabled   types.Bool   `tfsdk:"enabled"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	Sender    types.String `tfsdk:"sender"`
	UseAuth   types.Bool   `tfsdk:"use_auth"`
	UseSender types.Bool   `tfsdk:"use_sender"`
	UseSsl    types.Bool   `tfsdk:"use_ssl"`
	Username  types.String `tfsdk:"username"`
	XPassword types.String `tfsdk:"x_password"`
}

func (d *superSMTPModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperSmtp{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// The SMTP server fields only apply when SMTP is enabled.
	if d.Enabled.ValueBool() {
		if !ut.IsEmptyString(d.Host) {
			model.Host = d.Host.ValueString()
		}
		if !d.Port.IsNull() && !d.Port.IsUnknown() {
			model.Port = int(d.Port.ValueInt32())
		}
		if !ut.IsEmptyString(d.Sender) {
			model.Sender = d.Sender.ValueString()
		}
		if !ut.IsEmptyString(d.Username) {
			model.Username = d.Username.ValueString()
		}
		if !ut.IsEmptyString(d.XPassword) {
			model.XPassword = d.XPassword.ValueString()
		}
		model.UseAuth = d.UseAuth.ValueBool()
		model.UseSender = d.UseSender.ValueBool()
		model.UseSsl = d.UseSsl.ValueBool()
	}

	return model, diags
}

func (d *superSMTPModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperSmtp)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperSmtp")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// The SMTP server fields are only meaningful when SMTP is enabled.
	if model.Enabled {
		d.Host = ut.StringOrNull(model.Host)
		d.Port = ut.Int32OrNull(model.Port)
		d.Sender = ut.StringOrNull(model.Sender)
		d.UseAuth = types.BoolValue(model.UseAuth)
		d.UseSender = types.BoolValue(model.UseSender)
		d.UseSsl = types.BoolValue(model.UseSsl)
		d.Username = ut.StringOrNull(model.Username)
		d.XPassword = ut.StringOrNull(model.XPassword)
	} else {
		d.Host = types.StringNull()
		d.Port = types.Int32Null()
		d.Sender = types.StringNull()
		d.UseAuth = types.BoolNull()
		d.UseSender = types.BoolNull()
		d.UseSsl = types.BoolNull()
		d.Username = types.StringNull()
		d.XPassword = types.StringNull()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &superSMTPModel{}
	_ resource.Resource                     = &superSMTPResource{}
	_ resource.ResourceWithConfigure        = &superSMTPResource{}
	_ resource.ResourceWithImportState      = &superSMTPResource{}
	_ resource.ResourceWithConfigValidators = &superSMTPResource{}
)

type superSMTPResource struct {
	*base.GenericResource[*superSMTPModel]
}

func (r *superSMTPResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false),
			path.MatchRoot("host"),
			path.MatchRoot("port"),
			path.MatchRoot("sender"),
			path.MatchRoot("use_auth"),
			path.MatchRoot("use_sender"),
			path.MatchRoot("use_ssl"),
			path.MatchRoot("username"),
			path.MatchRoot("x_password"),
		),
	}
}

func (r *superSMTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_smtp` resource manages the custom SMTP server used by a UniFi console to " +
			"send email notifications. The server attributes only apply when `enabled` is `true`.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the custom SMTP server is enabled.",
				Required:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The SMTP server hostname or IP address. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "The SMTP server port. Valid values are 1-65535. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"sender": schema.StringAttribute{
				MarkdownDescription: "The email address used as the sender (From) for outgoing messages. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"use_auth": schema.BoolAttribute{
				MarkdownDescription: "Whether the SMTP server requires authentication. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"use_sender": schema.BoolAttribute{
				MarkdownDescription: "Whether to use a custom sender (From) address. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL/TLS when connecting to the SMTP server. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username for SMTP authentication. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"x_password": schema.StringAttribute{
				MarkdownDescription: "The password for SMTP authentication. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// NewSuperSMTPResource creates a new instance of the super SMTP setting resource.
func NewSuperSMTPResource() resource.Resource {
	r := &superSMTPResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_smtp",
		func() *superSMTPModel { return &superSMTPModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperSmtp(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperSmtp)
			return client.UpdateSettingSuperSmtp(ctx, site, b)
		},
	)
	return r
}
