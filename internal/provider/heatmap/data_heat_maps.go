package heatmap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// HeatMapsDatasourceModel represents the data model for a list of UniFi heat maps.
type HeatMapsDatasourceModel struct {
	Site     types.String             `tfsdk:"site"`
	HeatMaps []HeatMapDatasourceModel `tfsdk:"heat_maps"`
}

func (m *HeatMapsDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *HeatMapsDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *HeatMapsDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &heatMapsDatasource{}
	_ datasource.DataSourceWithConfigure = &heatMapsDatasource{}
	_ base.Resource                      = &heatMapsDatasource{}
)

type heatMapsDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHeatMapsDatasource() datasource.DataSource {
	return &heatMapsDatasource{}
}

func (d *heatMapsDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *heatMapsDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *heatMapsDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *heatMapsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *heatMapsDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_heat_maps"
}

func (d *heatMapsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_heat_maps` data source retrieves all heat maps configured on a site.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"heat_maps": schema.ListNestedAttribute{
				MarkdownDescription: "The list of heat maps.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          ut.ID(),
						"site":        ut.SiteAttribute(),
						"map_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "The identifier of the map this heat map belongs to."},
						"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "The name of the heat map."},
						"description": schema.StringAttribute{Computed: true, MarkdownDescription: "A free-form description of the heat map."},
						"type":        schema.StringAttribute{Computed: true, MarkdownDescription: "The metric the heat map represents."},
					},
				},
			},
		},
	}
}

func (d *heatMapsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HeatMapsDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	heatMaps, err := d.client.ListHeatMap(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list heat maps", err.Error())
		return
	}

	state.HeatMaps = make([]HeatMapDatasourceModel, 0, len(heatMaps))
	for i := range heatMaps {
		var item HeatMapDatasourceModel
		resp.Diagnostics.Append(item.Merge(ctx, &heatMaps[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.SetSite(site)
		state.HeatMaps = append(state.HeatMaps, item)
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
