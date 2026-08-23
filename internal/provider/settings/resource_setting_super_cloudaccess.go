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

// superCloudaccessModel represents the UniFi cloud access settings, including
// the AWS IoT certificate material used to establish the cloud tunnel.
type superCloudaccessModel struct {
	base.Model
	Enabled         types.Bool   `tfsdk:"enabled"`
	DeviceID        types.String `tfsdk:"device_id"`
	UbicUUID        types.String `tfsdk:"ubic_uuid"`
	DeviceAuth      types.String `tfsdk:"device_auth"`
	XCertificateArn types.String `tfsdk:"x_certificate_arn"`
	XCertificatePem types.String `tfsdk:"x_certificate_pem"`
	XPrivateKey     types.String `tfsdk:"x_private_key"`
}

func (d *superCloudaccessModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperCloudaccess{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// device_id and ubic_uuid are not gated - always reflect them.
	if !ut.IsEmptyString(d.DeviceID) {
		model.DeviceID = d.DeviceID.ValueString()
	}
	if !ut.IsEmptyString(d.UbicUUID) {
		model.UbicUuid = d.UbicUUID.ValueString()
	}

	// The credential fields only apply when cloud access is enabled.
	if d.Enabled.ValueBool() {
		if !ut.IsEmptyString(d.DeviceAuth) {
			model.DeviceAuth = d.DeviceAuth.ValueString()
		}
		if !ut.IsEmptyString(d.XCertificateArn) {
			model.XCertificateArn = d.XCertificateArn.ValueString()
		}
		if !ut.IsEmptyString(d.XCertificatePem) {
			model.XCertificatePem = d.XCertificatePem.ValueString()
		}
		if !ut.IsEmptyString(d.XPrivateKey) {
			model.XPrivateKey = d.XPrivateKey.ValueString()
		}
	}

	return model, diags
}

func (d *superCloudaccessModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperCloudaccess)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperCloudaccess")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// device_id and ubic_uuid are always reflected.
	d.DeviceID = ut.StringOrNull(model.DeviceID)
	d.UbicUUID = ut.StringOrNull(model.UbicUuid)

	// The credential fields are only meaningful when cloud access is enabled.
	if model.Enabled {
		d.DeviceAuth = ut.StringOrNull(model.DeviceAuth)
		d.XCertificateArn = ut.StringOrNull(model.XCertificateArn)
		d.XCertificatePem = ut.StringOrNull(model.XCertificatePem)
		d.XPrivateKey = ut.StringOrNull(model.XPrivateKey)
	} else {
		d.DeviceAuth = types.StringNull()
		d.XCertificateArn = types.StringNull()
		d.XCertificatePem = types.StringNull()
		d.XPrivateKey = types.StringNull()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &superCloudaccessModel{}
	_ resource.Resource                     = &superCloudaccessResource{}
	_ resource.ResourceWithConfigure        = &superCloudaccessResource{}
	_ resource.ResourceWithImportState      = &superCloudaccessResource{}
	_ resource.ResourceWithConfigValidators = &superCloudaccessResource{}
)

type superCloudaccessResource struct {
	*base.GenericResource[*superCloudaccessModel]
}

func (r *superCloudaccessResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false),
			path.MatchRoot("device_auth"),
			path.MatchRoot("x_certificate_arn"),
			path.MatchRoot("x_certificate_pem"),
			path.MatchRoot("x_private_key"),
		),
	}
}

func (r *superCloudaccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_cloudaccess` resource manages remote cloud access for a UniFi console, " +
			"including the AWS IoT certificate material used to establish the cloud tunnel. The credential attributes only " +
			"apply when `enabled` is `true`.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether remote cloud access is enabled for the console.",
				Required:            true,
			},
			"device_id": schema.StringAttribute{
				MarkdownDescription: "The device identifier registered with the Ubiquiti cloud.",
				Optional:            true,
				Computed:            true,
			},
			"ubic_uuid": schema.StringAttribute{
				MarkdownDescription: "The Ubiquiti cloud (UBIC) UUID associated with the console.",
				Optional:            true,
				Computed:            true,
			},
			"device_auth": schema.StringAttribute{
				MarkdownDescription: "The device authentication secret used for the cloud tunnel. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"x_certificate_arn": schema.StringAttribute{
				MarkdownDescription: "The AWS IoT certificate ARN for the cloud tunnel. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
			},
			"x_certificate_pem": schema.StringAttribute{
				MarkdownDescription: "The AWS IoT certificate (PEM) for the cloud tunnel. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"x_private_key": schema.StringAttribute{
				MarkdownDescription: "The AWS IoT private key for the cloud tunnel. Only applicable when `enabled` is `true`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// NewSuperCloudaccessResource creates a new instance of the super cloud access setting resource.
func NewSuperCloudaccessResource() resource.Resource {
	r := &superCloudaccessResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_cloudaccess",
		func() *superCloudaccessModel { return &superCloudaccessModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperCloudaccess(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperCloudaccess)
			return client.UpdateSettingSuperCloudaccess(ctx, site, b)
		},
	)
	return r
}
