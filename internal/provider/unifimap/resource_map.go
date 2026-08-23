package unifimap

import (
	"context"
	"fmt"

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

// MapModel represents the data model for a UniFi map (floor plan).
type MapModel struct {
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
func (m *MapModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.Map{
		ID:         m.ID.ValueString(),
		Lat:        m.Lat.ValueString(),
		Lng:        m.Lng.ValueString(),
		MapTypeID:  m.MapTypeID.ValueString(),
		Name:       m.Name.ValueString(),
		OffsetLeft: m.OffsetLeft.ValueFloat64(),
		OffsetTop:  m.OffsetTop.ValueFloat64(),
		Opacity:    m.Opacity.ValueFloat64(),
		Selected:   m.Selected.ValueBool(),
		Tilt:       int(m.Tilt.ValueInt32()),
		Type:       m.Type.ValueString(),
		Unit:       m.Unit.ValueString(),
		Upp:        m.Upp.ValueFloat64(),
		Zoom:       int(m.Zoom.ValueInt32()),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *MapModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
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
	_ resource.Resource                = &mapResource{}
	_ resource.ResourceWithConfigure   = &mapResource{}
	_ resource.ResourceWithImportState = &mapResource{}
	_ base.Resource                    = &mapResource{}
	_ base.ResourceModel               = &MapModel{}
)

type mapResource struct {
	*base.GenericResource[*MapModel]
}

// NewMapResource creates a new instance of the map resource.
func NewMapResource() resource.Resource {
	return &mapResource{
		GenericResource: base.NewGenericResource(
			"unifi_map",
			func() *MapModel { return &MapModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetMap(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					m, _ := model.(*unifi.Map)
					return client.CreateMap(ctx, site, m)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					m, _ := model.(*unifi.Map)
					return client.UpdateMap(ctx, site, m)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteMap(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *mapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_map` resource manages a map (floor plan) in the UniFi controller. " +
			"Maps provide the spatial backdrop used to place devices and overlay heat maps (`unifi_heat_map`).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"lat": schema.StringAttribute{
				MarkdownDescription: "The latitude of the map centre (used by Google maps).",
				Optional:            true,
				Computed:            true,
			},
			"lng": schema.StringAttribute{
				MarkdownDescription: "The longitude of the map centre (used by Google maps).",
				Optional:            true,
				Computed:            true,
			},
			"map_type_id": schema.StringAttribute{
				MarkdownDescription: "The Google map type. One of `satellite`, `roadmap`, `hybrid` or `terrain`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("satellite", "roadmap", "hybrid", "terrain"),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the map.",
				Optional:            true,
				Computed:            true,
			},
			"offset_left": schema.Float64Attribute{
				MarkdownDescription: "The horizontal offset of the map image, in pixels.",
				Optional:            true,
				Computed:            true,
			},
			"offset_top": schema.Float64Attribute{
				MarkdownDescription: "The vertical offset of the map image, in pixels.",
				Optional:            true,
				Computed:            true,
			},
			"opacity": schema.Float64Attribute{
				MarkdownDescription: "The opacity of the map image, between `0` and `1`.",
				Optional:            true,
				Computed:            true,
			},
			"selected": schema.BoolAttribute{
				MarkdownDescription: "Whether this map is the currently selected map.",
				Optional:            true,
				Computed:            true,
			},
			"tilt": schema.Int32Attribute{
				MarkdownDescription: "The tilt angle of the map, in degrees.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of map. One of `designerMap`, `imageMap` or `googleMap`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("designerMap", "imageMap", "googleMap"),
				},
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "The distance unit of the map. One of `m` (metric) or `f` (imperial).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("m", "f"),
				},
			},
			"upp": schema.Float64Attribute{
				MarkdownDescription: "The units-per-pixel scale of the map.",
				Optional:            true,
				Computed:            true,
			},
			"zoom": schema.Int32Attribute{
				MarkdownDescription: "The zoom level of the map.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
