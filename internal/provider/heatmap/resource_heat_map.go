package heatmap

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

// HeatMapModel represents the data model for a UniFi heat map.
type HeatMapModel struct {
	base.Model
	MapID       types.String `tfsdk:"map_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HeatMapModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.HeatMap{
		ID:          m.ID.ValueString(),
		MapID:       m.MapID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        m.Type.ValueString(),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HeatMapModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.HeatMap)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.HeatMap, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.MapID = types.StringValue(model.MapID)
	m.Name = ut.StringOrNull(model.Name)
	m.Description = ut.StringOrNull(model.Description)
	m.Type = ut.StringOrNull(model.Type)
	return diags
}

var (
	_ resource.Resource                = &heatMapResource{}
	_ resource.ResourceWithConfigure   = &heatMapResource{}
	_ resource.ResourceWithImportState = &heatMapResource{}
	_ base.Resource                    = &heatMapResource{}
	_ base.ResourceModel               = &HeatMapModel{}
)

type heatMapResource struct {
	*base.GenericResource[*HeatMapModel]
}

// NewHeatMapResource creates a new instance of the heat map resource.
func NewHeatMapResource() resource.Resource {
	return &heatMapResource{
		GenericResource: base.NewGenericResource(
			"unifi_heat_map",
			func() *HeatMapModel { return &HeatMapModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetHeatMap(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					heatMap, _ := model.(*unifi.HeatMap)
					return client.CreateHeatMap(ctx, site, heatMap)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					heatMap, _ := model.(*unifi.HeatMap)
					return client.UpdateHeatMap(ctx, site, heatMap)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteHeatMap(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *heatMapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_heat_map` resource manages a heat map overlay attached to a floor " +
			"plan (`unifi_map`) in the UniFi controller. Heat maps visualise measured download or upload " +
			"performance across a mapped area.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"map_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the `unifi_map` this heat map belongs to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the heat map.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A free-form description of the heat map.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The metric the heat map represents. One of `download` or `upload`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("download", "upload"),
				},
			},
		},
	}
}
