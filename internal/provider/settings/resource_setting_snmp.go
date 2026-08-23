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

// snmpModel represents SNMP (Simple Network Management Protocol) agent settings
// for a UniFi site.
type snmpModel struct {
	base.Model
	Community types.String `tfsdk:"community"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	EnabledV3 types.Bool   `tfsdk:"enabled_v3"`
	Username  types.String `tfsdk:"username"`
	XPassword types.String `tfsdk:"x_password"`
}

func (d *snmpModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSnmp{
		ID:        d.ID.ValueString(),
		Enabled:   d.Enabled.ValueBool(),
		EnabledV3: d.EnabledV3.ValueBool(),
	}
	if !ut.IsEmptyString(d.Community) {
		model.Community = d.Community.ValueString()
	}
	if !ut.IsEmptyString(d.Username) {
		model.Username = d.Username.ValueString()
	}
	if !ut.IsEmptyString(d.XPassword) {
		model.XPassword = d.XPassword.ValueString()
	}

	return model, diags
}

func (d *snmpModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSnmp)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSnmp")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)
	d.EnabledV3 = types.BoolValue(model.EnabledV3)
	d.Community = ut.StringOrNull(model.Community)
	d.Username = ut.StringOrNull(model.Username)
	d.XPassword = ut.StringOrNull(model.XPassword)

	return diags
}

var (
	_ base.ResourceModel               = &snmpModel{}
	_ resource.Resource                = &snmpResource{}
	_ resource.ResourceWithConfigure   = &snmpResource{}
	_ resource.ResourceWithImportState = &snmpResource{}
)

type snmpResource struct {
	*base.GenericResource[*snmpModel]
}

func (r *snmpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_snmp` resource manages the SNMP (Simple Network Management Protocol) agent " +
			"settings for a UniFi site, exposing device metrics to SNMP monitoring systems. Both SNMP v1/v2c (community-based) " +
			"and SNMP v3 (user/password-based) can be enabled.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the SNMP v1/v2c agent is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"enabled_v3": schema.BoolAttribute{
				MarkdownDescription: "Whether the SNMP v3 agent is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"community": schema.StringAttribute{
				MarkdownDescription: "SNMP v1/v2c community string used by monitoring systems to read metrics.",
				Optional:            true,
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "SNMP v3 username.",
				Optional:            true,
				Computed:            true,
			},
			"x_password": schema.StringAttribute{
				MarkdownDescription: "SNMP v3 authentication password.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// NewSnmpResource creates a new instance of the SNMP setting resource.
func NewSnmpResource() resource.Resource {
	r := &snmpResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_snmp",
		func() *snmpModel { return &snmpModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSnmp(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSnmp)
			return client.UpdateSettingSnmp(ctx, site, b)
		},
	)
	return r
}
