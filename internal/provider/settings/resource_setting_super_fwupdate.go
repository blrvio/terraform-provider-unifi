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

// superFwupdateModel represents the firmware and controller update channel
// settings for a UniFi console.
type superFwupdateModel struct {
	base.Model
	ControllerChannel types.String `tfsdk:"controller_channel"`
	FirmwareChannel   types.String `tfsdk:"firmware_channel"`
	SsoEnabled        types.Bool   `tfsdk:"sso_enabled"`
}

func (d *superFwupdateModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperFwupdate{
		ID:         d.ID.ValueString(),
		SsoEnabled: d.SsoEnabled.ValueBool(),
	}
	if !ut.IsEmptyString(d.ControllerChannel) {
		model.ControllerChannel = d.ControllerChannel.ValueString()
	}
	if !ut.IsEmptyString(d.FirmwareChannel) {
		model.FirmwareChannel = d.FirmwareChannel.ValueString()
	}

	return model, diags
}

func (d *superFwupdateModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperFwupdate)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperFwupdate")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.ControllerChannel = ut.StringOrNull(model.ControllerChannel)
	d.FirmwareChannel = ut.StringOrNull(model.FirmwareChannel)
	d.SsoEnabled = types.BoolValue(model.SsoEnabled)

	return diags
}

var (
	_ base.ResourceModel               = &superFwupdateModel{}
	_ resource.Resource                = &superFwupdateResource{}
	_ resource.ResourceWithConfigure   = &superFwupdateResource{}
	_ resource.ResourceWithImportState = &superFwupdateResource{}
)

type superFwupdateResource struct {
	*base.GenericResource[*superFwupdateModel]
}

func (r *superFwupdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	channelValidators := func() []validator.String {
		return []validator.String{
			stringvalidator.OneOf("internal", "alpha", "beta", "release-candidate", "release"),
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_fwupdate` resource manages the update channels used by a UniFi console " +
			"for controller and device firmware releases.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"controller_channel": schema.StringAttribute{
				MarkdownDescription: "The release channel used for controller (application) updates. Valid values are " +
					"`internal`, `alpha`, `beta`, `release-candidate`, or `release`.",
				Optional:   true,
				Computed:   true,
				Validators: channelValidators(),
			},
			"firmware_channel": schema.StringAttribute{
				MarkdownDescription: "The release channel used for device firmware updates. Valid values are " +
					"`internal`, `alpha`, `beta`, `release-candidate`, or `release`.",
				Optional:   true,
				Computed:   true,
				Validators: channelValidators(),
			},
			"sso_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Single Sign-On (SSO) with the Ubiquiti account is enabled for firmware updates.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewSuperFwupdateResource creates a new instance of the super firmware update setting resource.
func NewSuperFwupdateResource() resource.Resource {
	r := &superFwupdateResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_fwupdate",
		func() *superFwupdateModel { return &superFwupdateModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperFwupdate(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperFwupdate)
			return client.UpdateSettingSuperFwupdate(ctx, site, b)
		},
	)
	return r
}
