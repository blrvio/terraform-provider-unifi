package dashboard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// DashboardsDatasourceModel represents the data model for a list of UniFi dashboards.
type DashboardsDatasourceModel struct {
	Site       types.String               `tfsdk:"site"`
	Dashboards []DashboardDatasourceModel `tfsdk:"dashboards"`
}

func (m *DashboardsDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *DashboardsDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *DashboardsDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &dashboardsDatasource{}
	_ datasource.DataSourceWithConfigure = &dashboardsDatasource{}
	_ base.Resource                      = &dashboardsDatasource{}
)

type dashboardsDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewDashboardsDatasource() datasource.DataSource {
	return &dashboardsDatasource{}
}

func (d *dashboardsDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *dashboardsDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *dashboardsDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *dashboardsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *dashboardsDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_dashboards"
}

func (d *dashboardsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dashboards` data source retrieves all dashboards configured on a site.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"dashboards": schema.ListNestedAttribute{
				MarkdownDescription: "The list of dashboards.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 ut.ID(),
						"site":               ut.SiteAttribute(),
						"name":               schema.StringAttribute{Computed: true, MarkdownDescription: "The name of the dashboard."},
						"desc":               schema.StringAttribute{Computed: true, MarkdownDescription: "A free-form description of the dashboard."},
						"controller_version": schema.StringAttribute{Computed: true, MarkdownDescription: "The controller version the dashboard was authored against."},
						"is_public":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the dashboard is publicly visible."},
						"modules": schema.ListNestedAttribute{
							MarkdownDescription: "The modules that make up the dashboard.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: dashboardModuleDatasourceAttributes(),
							},
						},
					},
				},
			},
		},
	}
}

func (d *dashboardsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DashboardsDatasourceModel
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

	state.Dashboards = make([]DashboardDatasourceModel, 0, len(dashboards))
	for i := range dashboards {
		var item DashboardDatasourceModel
		resp.Diagnostics.Append(item.Merge(ctx, &dashboards[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.SetSite(site)
		state.Dashboards = append(state.Dashboards, item)
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
