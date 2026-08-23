package heatmappoint

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

// HeatMapPointDatasourceModel represents the data model for a UniFi heat map point data source.
type HeatMapPointDatasourceModel struct {
	base.Model
	HeatmapID     types.String  `tfsdk:"heatmap_id"`
	DownloadSpeed types.Float64 `tfsdk:"download_speed"`
	UploadSpeed   types.Float64 `tfsdk:"upload_speed"`
	X             types.Float64 `tfsdk:"x"`
	Y             types.Float64 `tfsdk:"y"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HeatMapPointDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HeatMapPointDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.HeatMapPoint)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.HeatMapPoint, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.HeatmapID = types.StringValue(model.HeatmapID)
	m.DownloadSpeed = ut.Float64OrNull(model.DownloadSpeed)
	m.UploadSpeed = ut.Float64OrNull(model.UploadSpeed)
	m.X = ut.Float64OrNull(model.X)
	m.Y = ut.Float64OrNull(model.Y)
	return diags
}

var (
	_ datasource.DataSource              = &heatMapPointDatasource{}
	_ datasource.DataSourceWithConfigure = &heatMapPointDatasource{}
	_ base.Resource                      = &heatMapPointDatasource{}
)

type heatMapPointDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHeatMapPointDatasource() datasource.DataSource {
	return &heatMapPointDatasource{}
}

func (d *heatMapPointDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *heatMapPointDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *heatMapPointDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *heatMapPointDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *heatMapPointDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_heat_map_point"
}

func (d *heatMapPointDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_heat_map_point` data source retrieves an existing heat map point by id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"heatmap_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the `unifi_heat_map` this point belongs to.",
				Computed:            true,
			},
			"download_speed": schema.Float64Attribute{
				MarkdownDescription: "The measured download speed at this point.",
				Computed:            true,
			},
			"upload_speed": schema.Float64Attribute{
				MarkdownDescription: "The measured upload speed at this point.",
				Computed:            true,
			},
			"x": schema.Float64Attribute{
				MarkdownDescription: "The horizontal position of the point on the map.",
				Computed:            true,
			},
			"y": schema.Float64Attribute{
				MarkdownDescription: "The vertical position of the point on the map.",
				Computed:            true,
			},
		},
	}
}

func (d *heatMapPointDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HeatMapPointDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	heatMapPoints, err := d.client.ListHeatMapPoint(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list heat map points", err.Error())
		return
	}

	id := state.GetID()
	if id == "" {
		resp.Diagnostics.AddError("Missing lookup key", "`id` must be set to look up a heat map point.")
		return
	}

	var found *unifi.HeatMapPoint
	for i := range heatMapPoints {
		h := heatMapPoints[i]
		if h.ID == id {
			found = &h
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Heat map point not found", fmt.Sprintf("No heat map point matching id=%q was found", id))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
