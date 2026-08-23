package heatmappoint

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

// HeatMapPointModel represents the data model for a UniFi heat map point.
type HeatMapPointModel struct {
	base.Model
	HeatmapID     types.String  `tfsdk:"heatmap_id"`
	DownloadSpeed types.Float64 `tfsdk:"download_speed"`
	UploadSpeed   types.Float64 `tfsdk:"upload_speed"`
	X             types.Float64 `tfsdk:"x"`
	Y             types.Float64 `tfsdk:"y"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HeatMapPointModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.HeatMapPoint{
		ID:            m.ID.ValueString(),
		HeatmapID:     m.HeatmapID.ValueString(),
		DownloadSpeed: m.DownloadSpeed.ValueFloat64(),
		UploadSpeed:   m.UploadSpeed.ValueFloat64(),
		X:             m.X.ValueFloat64(),
		Y:             m.Y.ValueFloat64(),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HeatMapPointModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
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
	_ resource.Resource                = &heatMapPointResource{}
	_ resource.ResourceWithConfigure   = &heatMapPointResource{}
	_ resource.ResourceWithImportState = &heatMapPointResource{}
	_ base.Resource                    = &heatMapPointResource{}
	_ base.ResourceModel               = &HeatMapPointModel{}
)

type heatMapPointResource struct {
	*base.GenericResource[*HeatMapPointModel]
}

// NewHeatMapPointResource creates a new instance of the heat map point resource.
func NewHeatMapPointResource() resource.Resource {
	return &heatMapPointResource{
		GenericResource: base.NewGenericResource(
			"unifi_heat_map_point",
			func() *HeatMapPointModel { return &HeatMapPointModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetHeatMapPoint(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					heatMapPoint, _ := model.(*unifi.HeatMapPoint)
					return client.CreateHeatMapPoint(ctx, site, heatMapPoint)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					heatMapPoint, _ := model.(*unifi.HeatMapPoint)
					return client.UpdateHeatMapPoint(ctx, site, heatMapPoint)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteHeatMapPoint(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *heatMapPointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_heat_map_point` resource manages a single measurement point on a " +
			"heat map (`unifi_heat_map`) in the UniFi controller. Each point records the measured download " +
			"and upload performance at a specific position on the mapped area.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"heatmap_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the `unifi_heat_map` this point belongs to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"download_speed": schema.Float64Attribute{
				MarkdownDescription: "The measured download speed at this point.",
				Optional:            true,
				Computed:            true,
			},
			"upload_speed": schema.Float64Attribute{
				MarkdownDescription: "The measured upload speed at this point.",
				Optional:            true,
				Computed:            true,
			},
			"x": schema.Float64Attribute{
				MarkdownDescription: "The horizontal position of the point on the map.",
				Optional:            true,
				Computed:            true,
			},
			"y": schema.Float64Attribute{
				MarkdownDescription: "The vertical position of the point on the map.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
