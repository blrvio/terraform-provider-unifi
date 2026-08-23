package unifimap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// MapsDatasourceModel represents the data model for a list of UniFi maps.
type MapsDatasourceModel struct {
	Site types.String         `tfsdk:"site"`
	Maps []MapDatasourceModel `tfsdk:"maps"`
}

func (m *MapsDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *MapsDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *MapsDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &mapsDatasource{}
	_ datasource.DataSourceWithConfigure = &mapsDatasource{}
	_ base.Resource                      = &mapsDatasource{}
)

type mapsDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewMapsDatasource() datasource.DataSource {
	return &mapsDatasource{}
}

func (d *mapsDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *mapsDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *mapsDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *mapsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *mapsDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_maps"
}

func (d *mapsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_maps` data source retrieves all maps (floor plans) configured on a site.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"maps": schema.ListNestedAttribute{
				MarkdownDescription: "The list of maps.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          ut.ID(),
						"site":        ut.SiteAttribute(),
						"lat":         schema.StringAttribute{Computed: true, MarkdownDescription: "The latitude of the map centre."},
						"lng":         schema.StringAttribute{Computed: true, MarkdownDescription: "The longitude of the map centre."},
						"map_type_id": schema.StringAttribute{Computed: true, MarkdownDescription: "The Google map type."},
						"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "The name of the map."},
						"offset_left": schema.Float64Attribute{Computed: true, MarkdownDescription: "The horizontal offset of the map image, in pixels."},
						"offset_top":  schema.Float64Attribute{Computed: true, MarkdownDescription: "The vertical offset of the map image, in pixels."},
						"opacity":     schema.Float64Attribute{Computed: true, MarkdownDescription: "The opacity of the map image."},
						"selected":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this map is the currently selected map."},
						"tilt":        schema.Int32Attribute{Computed: true, MarkdownDescription: "The tilt angle of the map, in degrees."},
						"type":        schema.StringAttribute{Computed: true, MarkdownDescription: "The type of map."},
						"unit":        schema.StringAttribute{Computed: true, MarkdownDescription: "The distance unit of the map."},
						"upp":         schema.Float64Attribute{Computed: true, MarkdownDescription: "The units-per-pixel scale of the map."},
						"zoom":        schema.Int32Attribute{Computed: true, MarkdownDescription: "The zoom level of the map."},
					},
				},
			},
		},
	}
}

func (d *mapsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MapsDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	maps, err := d.client.ListMap(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list maps", err.Error())
		return
	}

	state.Maps = make([]MapDatasourceModel, 0, len(maps))
	for i := range maps {
		var item MapDatasourceModel
		resp.Diagnostics.Append(item.Merge(ctx, &maps[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.SetSite(site)
		state.Maps = append(state.Maps, item)
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
