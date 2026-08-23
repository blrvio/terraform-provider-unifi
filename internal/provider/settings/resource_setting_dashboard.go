package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// dashboardWidgetModel represents a single widget's visibility on the UniFi
// Network dashboard.
type dashboardWidgetModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Name    types.String `tfsdk:"name"`
}

func (m *dashboardWidgetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"name":    types.StringType,
	}
}

// dashboardModel represents the UniFi Network dashboard layout settings for a
// site.
type dashboardModel struct {
	base.Model
	LayoutPreference types.String `tfsdk:"layout_preference"`
	Widgets          types.List   `tfsdk:"widgets"`
}

func (d *dashboardModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingDashboard{
		ID:               d.ID.ValueString(),
		LayoutPreference: d.LayoutPreference.ValueString(),
	}

	if ut.IsDefined(d.Widgets) {
		var items []dashboardWidgetModel
		diags.Append(d.Widgets.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.Widgets = make([]unifi.SettingDashboardWidgets, 0, len(items))
		for _, it := range items {
			model.Widgets = append(model.Widgets, unifi.SettingDashboardWidgets{
				Enabled: it.Enabled.ValueBool(),
				Name:    it.Name.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *dashboardModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingDashboard)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingDashboard")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.LayoutPreference = ut.StringOrNull(model.LayoutPreference)

	widgetItems := make([]dashboardWidgetModel, 0, len(model.Widgets))
	for _, it := range model.Widgets {
		widgetItems = append(widgetItems, dashboardWidgetModel{
			Enabled: types.BoolValue(it.Enabled),
			Name:    types.StringValue(it.Name),
		})
	}
	widgetList, widgetDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&dashboardWidgetModel{}).AttributeTypes()}, widgetItems)
	diags.Append(widgetDiags...)
	if diags.HasError() {
		return diags
	}
	d.Widgets = widgetList

	return diags
}

var (
	_ base.ResourceModel               = &dashboardModel{}
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

type dashboardResource struct {
	*base.GenericResource[*dashboardModel]
}

func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_dashboard` resource manages the UniFi Network dashboard layout for a site. " +
			"`layout_preference` chooses between the controller's automatic arrangement and a manually curated one, and " +
			"`widgets` toggles the visibility of individual dashboard cards (WiFi metrics, most-active apps/clients, WAN " +
			"activity, and so on).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"layout_preference": schema.StringAttribute{
				MarkdownDescription: "Dashboard layout mode. `auto` lets the controller arrange widgets, while `manual` " +
					"honors the explicit `widgets` selection.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},
			"widgets": schema.ListNestedAttribute{
				MarkdownDescription: "Per-widget visibility on the dashboard, applied when `layout_preference` is `manual`.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether this widget is shown on the dashboard.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Identifier of the dashboard widget.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"critical_traffic_prioritization", "cybersecure", "traffic_identification",
									"wifi_technology", "wifi_channels", "wifi_client_experience", "wifi_tx_retries",
									"most_active_apps_aps_clients", "most_active_apps_clients", "most_active_aps_clients",
									"most_active_apps_aps", "most_active_apps", "v2_most_active_aps",
									"v2_most_active_clients", "wifi_connectivity", "ap_radio_density",
									"wifi_channel_preset_configuration", "most_common_client_fingerprints",
									"wan_activity",
								),
							},
						},
					},
				},
			},
		},
	}
}

// NewDashboardResource creates a new instance of the dashboard setting resource.
func NewDashboardResource() resource.Resource {
	r := &dashboardResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_dashboard",
		func() *dashboardModel { return &dashboardModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingDashboard(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingDashboard)
			return client.UpdateSettingDashboard(ctx, site, b)
		},
	)
	return r
}
