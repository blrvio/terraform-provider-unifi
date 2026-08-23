package unifimap

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

// MapDatasourceModel represents the data model for a UniFi map data source.
type MapDatasourceModel struct {
	base.Model
	Lat        types.String  `tfsdk:"lat"`
	Lng        types.String  `tfsdk:"lng"`
	MapTypeID  types.String  `tfsdk:"map_type_id"`
	Name       types.String  `tfsdk:"name"`
	OffsetLeft types.Float64 `tfsdk:"offset_left"`
	OffsetTop  types.Float64 `tfsdk:"offset_top"`
	Opacity    types.Float64 `tfsdk:"opacity"`
	Selected   types.Bool    `tfsdk:"selected"`
	Tilt       types.Int32   `tfsdk:"tilt"`
	Type       types.String  `tfsdk:"type"`
	Unit       types.String  `tfsdk:"unit"`
	Upp        types.Float64 `tfsdk:"upp"`
	Zoom       types.Int32   `tfsdk:"zoom"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *MapDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *MapDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.Map)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Map, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Lat = ut.StringOrNull(model.Lat)
	m.Lng = ut.StringOrNull(model.Lng)
	m.MapTypeID = ut.StringOrNull(model.MapTypeID)
	m.Name = ut.StringOrNull(model.Name)
	m.OffsetLeft = ut.Float64OrNull(model.OffsetLeft)
	m.OffsetTop = ut.Float64OrNull(model.OffsetTop)
	m.Opacity = ut.Float64OrNull(model.Opacity)
	m.Selected = types.BoolValue(model.Selected)
	m.Tilt = ut.Int32OrNull(model.Tilt)
	m.Type = ut.StringOrNull(model.Type)
	m.Unit = ut.StringOrNull(model.Unit)
	m.Upp = ut.Float64OrNull(model.Upp)
	m.Zoom = ut.Int32OrNull(model.Zoom)
	return diags
}

var (
	_ datasource.DataSource              = &mapDatasource{}
	_ datasource.DataSourceWithConfigure = &mapDatasource{}
	_ base.Resource                      = &mapDatasource{}
)

type mapDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewMapDatasource() datasource.DataSource {
	return &mapDatasource{}
}

func (d *mapDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *mapDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *mapDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *mapDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *mapDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_map"
}

func (d *mapDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_map` data source retrieves an existing map (floor plan) by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"lat": schema.StringAttribute{
				MarkdownDescription: "The latitude of the map centre.",
				Computed:            true,
			},
			"lng": schema.StringAttribute{
				MarkdownDescription: "The longitude of the map centre.",
				Computed:            true,
			},
			"map_type_id": schema.StringAttribute{
				MarkdownDescription: "The Google map type (`satellite`, `roadmap`, `hybrid` or `terrain`).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the map to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"offset_left": schema.Float64Attribute{
				MarkdownDescription: "The horizontal offset of the map image, in pixels.",
				Computed:            true,
			},
			"offset_top": schema.Float64Attribute{
				MarkdownDescription: "The vertical offset of the map image, in pixels.",
				Computed:            true,
			},
			"opacity": schema.Float64Attribute{
				MarkdownDescription: "The opacity of the map image, between `0` and `1`.",
				Computed:            true,
			},
			"selected": schema.BoolAttribute{
				MarkdownDescription: "Whether this map is the currently selected map.",
				Computed:            true,
			},
			"tilt": schema.Int32Attribute{
				MarkdownDescription: "The tilt angle of the map, in degrees.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of map (`designerMap`, `imageMap` or `googleMap`).",
				Computed:            true,
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "The distance unit of the map (`m` or `f`).",
				Computed:            true,
			},
			"upp": schema.Float64Attribute{
				MarkdownDescription: "The units-per-pixel scale of the map.",
				Computed:            true,
			},
			"zoom": schema.Int32Attribute{
				MarkdownDescription: "The zoom level of the map.",
				Computed:            true,
			},
		},
	}
}

func (d *mapDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MapDatasourceModel
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

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a map.")
		return
	}

	var found *unifi.Map
	for i := range maps {
		mp := maps[i]
		if (id != "" && mp.ID == id) || (name != "" && mp.Name == name) {
			found = &mp
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Map not found", fmt.Sprintf("No map matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
