package dashboard

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

func dashboardModuleDatasourceAttributes() map[string]schema.Attribute {
	strAttr := func(desc string) schema.Attribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	return map[string]schema.Attribute{
		"id":           strAttr("The identifier of the module instance on the dashboard."),
		"module_id":    strAttr("The identifier of the module type."),
		"config":       strAttr("The module configuration payload (opaque JSON string)."),
		"restrictions": strAttr("The module restrictions payload (opaque JSON string)."),
	}
}

// DashboardDatasourceModel represents the data model for a dashboard data source.
type DashboardDatasourceModel struct {
	base.Model
	Name              types.String `tfsdk:"name"`
	Desc              types.String `tfsdk:"desc"`
	ControllerVersion types.String `tfsdk:"controller_version"`
	IsPublic          types.Bool   `tfsdk:"is_public"`
	Modules           types.List   `tfsdk:"modules"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *DashboardDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *DashboardDatasourceModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.Dashboard)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Dashboard, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = ut.StringOrNull(model.Name)
	m.Desc = ut.StringOrNull(model.Desc)
	m.ControllerVersion = ut.StringOrNull(model.ControllerVersion)
	m.IsPublic = types.BoolValue(model.IsPublic)

	list, listDiags := dashboardModulesToList(ctx, model.Modules)
	diags.Append(listDiags...)
	m.Modules = list
	return diags
}

var (
	_ datasource.DataSource              = &dashboardDatasource{}
	_ datasource.DataSourceWithConfigure = &dashboardDatasource{}
	_ base.Resource                      = &dashboardDatasource{}
)

type dashboardDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewDashboardDatasource() datasource.DataSource {
	return &dashboardDatasource{}
}

func (d *dashboardDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *dashboardDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *dashboardDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *dashboardDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *dashboardDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_dashboard"
}

func (d *dashboardDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dashboard` data source retrieves an existing dashboard by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the dashboard to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"desc": schema.StringAttribute{
				MarkdownDescription: "A free-form description of the dashboard.",
				Computed:            true,
			},
			"controller_version": schema.StringAttribute{
				MarkdownDescription: "The controller version the dashboard was authored against.",
				Computed:            true,
			},
			"is_public": schema.BoolAttribute{
				MarkdownDescription: "Whether the dashboard is publicly visible.",
				Computed:            true,
			},
			"modules": schema.ListNestedAttribute{
				MarkdownDescription: "The modules that make up the dashboard.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dashboardModuleDatasourceAttributes(),
				},
			},
		},
	}
}

func (d *dashboardDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DashboardDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	dashboards, err := d.client.ListDashboard(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list dashboards", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a dashboard.")
		return
	}

	var found *unifi.Dashboard
	for i := range dashboards {
		dash := dashboards[i]
		if (id != "" && dash.ID == id) || (name != "" && dash.Name == name) {
			found = &dash
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Dashboard not found", fmt.Sprintf("No dashboard matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
