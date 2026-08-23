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

// superSdnModel represents the UniFi SDN (cloud) account association settings
// for a console.
type superSdnModel struct {
	base.Model
	AuthToken       types.String `tfsdk:"auth_token"`
	DeviceID        types.String `tfsdk:"device_id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Migrated        types.Bool   `tfsdk:"migrated"`
	SsoLoginEnabled types.String `tfsdk:"sso_login_enabled"`
	UbicUUID        types.String `tfsdk:"ubic_uuid"`
}

func (d *superSdnModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperSdn{
		ID:       d.ID.ValueString(),
		Enabled:  d.Enabled.ValueBool(),
		Migrated: d.Migrated.ValueBool(),
	}
	if !ut.IsEmptyString(d.AuthToken) {
		model.AuthToken = d.AuthToken.ValueString()
	}
	if !ut.IsEmptyString(d.DeviceID) {
		model.DeviceID = d.DeviceID.ValueString()
	}
	if !ut.IsEmptyString(d.SsoLoginEnabled) {
		model.SsoLoginEnabled = d.SsoLoginEnabled.ValueString()
	}
	if !ut.IsEmptyString(d.UbicUUID) {
		model.UbicUuid = d.UbicUUID.ValueString()
	}

	return model, diags
}

func (d *superSdnModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperSdn)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperSdn")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)
	d.Migrated = types.BoolValue(model.Migrated)
	d.AuthToken = ut.StringOrNull(model.AuthToken)
	d.DeviceID = ut.StringOrNull(model.DeviceID)
	d.SsoLoginEnabled = ut.StringOrNull(model.SsoLoginEnabled)
	d.UbicUUID = ut.StringOrNull(model.UbicUuid)

	return diags
}

var (
	_ base.ResourceModel               = &superSdnModel{}
	_ resource.Resource                = &superSdnResource{}
	_ resource.ResourceWithConfigure   = &superSdnResource{}
	_ resource.ResourceWithImportState = &superSdnResource{}
)

type superSdnResource struct {
	*base.GenericResource[*superSdnModel]
}

func (r *superSdnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_sdn` resource manages the UniFi SDN (Ubiquiti cloud account) " +
			"association settings of a console.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the UniFi SDN cloud association is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"migrated": schema.BoolAttribute{
				MarkdownDescription: "Whether the console has been migrated to the UniFi SDN cloud account model.",
				Optional:            true,
				Computed:            true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "The authentication token used to associate the console with the Ubiquiti cloud account.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"device_id": schema.StringAttribute{
				MarkdownDescription: "The device identifier registered with the UniFi SDN cloud.",
				Optional:            true,
				Computed:            true,
			},
			"sso_login_enabled": schema.StringAttribute{
				MarkdownDescription: "Whether Single Sign-On (SSO) login is enabled for the SDN association.",
				Optional:            true,
				Computed:            true,
			},
			"ubic_uuid": schema.StringAttribute{
				MarkdownDescription: "The Ubiquiti cloud (UBIC) UUID associated with the console.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewSuperSdnResource creates a new instance of the super SDN setting resource.
func NewSuperSdnResource() resource.Resource {
	r := &superSdnResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_sdn",
		func() *superSdnModel { return &superSdnModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperSdn(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperSdn)
			return client.UpdateSettingSuperSdn(ctx, site, b)
		},
	)
	return r
}
