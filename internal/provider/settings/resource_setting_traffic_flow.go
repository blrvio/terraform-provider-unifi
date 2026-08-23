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

// trafficFlowModel represents the Traffic Flow settings for a UniFi site, which
// control which categories of traffic the controller is allowed to observe and
// manage.
type trafficFlowModel struct {
	base.Model
	EnabledAllowedTraffic        types.Bool `tfsdk:"enabled_allowed_traffic"`
	GatewayDNSEnabled            types.Bool `tfsdk:"gateway_dns_enabled"`
	UnifiDeviceManagementEnabled types.Bool `tfsdk:"unifi_device_management_enabled"`
	UnifiServicesEnabled         types.Bool `tfsdk:"unifi_services_enabled"`
}

func (d *trafficFlowModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingTrafficFlow{
		ID:                           d.ID.ValueString(),
		EnabledAllowedTraffic:        d.EnabledAllowedTraffic.ValueBool(),
		GatewayDNSEnabled:            d.GatewayDNSEnabled.ValueBool(),
		UnifiDeviceManagementEnabled: d.UnifiDeviceManagementEnabled.ValueBool(),
		UnifiServicesEnabled:         d.UnifiServicesEnabled.ValueBool(),
	}

	return model, diags
}

func (d *trafficFlowModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingTrafficFlow)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingTrafficFlow")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.EnabledAllowedTraffic = types.BoolValue(model.EnabledAllowedTraffic)
	d.GatewayDNSEnabled = types.BoolValue(model.GatewayDNSEnabled)
	d.UnifiDeviceManagementEnabled = types.BoolValue(model.UnifiDeviceManagementEnabled)
	d.UnifiServicesEnabled = types.BoolValue(model.UnifiServicesEnabled)

	return diags
}

var (
	_ base.ResourceModel               = &trafficFlowModel{}
	_ resource.Resource                = &trafficFlowResource{}
	_ resource.ResourceWithConfigure   = &trafficFlowResource{}
	_ resource.ResourceWithImportState = &trafficFlowResource{}
)

type trafficFlowResource struct {
	*base.GenericResource[*trafficFlowModel]
}

func (r *trafficFlowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_traffic_flow` resource manages Traffic Flow settings for a UniFi site, controlling which " +
			"categories of traffic the controller may observe and manage (allowed traffic, gateway DNS, UniFi device management, and UniFi services).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled_allowed_traffic": schema.BoolAttribute{
				MarkdownDescription: "Whether allowed-traffic flows are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"gateway_dns_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether gateway DNS traffic flows are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"unifi_device_management_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether UniFi device management traffic flows are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"unifi_services_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether UniFi services traffic flows are enabled.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewTrafficFlowResource creates a new instance of the Traffic Flow setting resource.
func NewTrafficFlowResource() resource.Resource {
	r := &trafficFlowResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_traffic_flow",
		func() *trafficFlowModel { return &trafficFlowModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingTrafficFlow(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingTrafficFlow)
			return client.UpdateSettingTrafficFlow(ctx, site, b)
		},
	)
	return r
}
