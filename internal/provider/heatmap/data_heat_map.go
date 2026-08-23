package heatmap

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

// HeatMapDatasourceModel represents the data model for a UniFi heat map data source.
type HeatMapDatasourceModel struct {
	base.Model
	MapID       types.String `tfsdk:"map_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HeatMapDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HeatMapDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
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
	_ datasource.DataSource              = &heatMapDatasource{}
	_ datasource.DataSourceWithConfigure = &heatMapDatasource{}
	_ base.Resource                      = &heatMapDatasource{}
)

type heatMapDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHeatMapDatasource() datasource.DataSource {
	return &heatMapDatasource{}
}

func (d *heatMapDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *heatMapDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *heatMapDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *heatMapDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *heatMapDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_heat_map"
}

func (d *heatMapDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_heat_map` data source retrieves an existing heat map by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"map_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the `unifi_map` this heat map belongs to.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the heat map to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A free-form description of the heat map.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The metric the heat map represents (`download` or `upload`).",
				Computed:            true,
			},
		},
	}
}

func (d *heatMapDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HeatMapDatasourceModel
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

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a heat map.")
		return
	}

	var found *unifi.HeatMap
	for i := range heatMaps {
		h := heatMaps[i]
		if (id != "" && h.ID == id) || (name != "" && h.Name == name) {
			found = &h
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Heat map not found", fmt.Sprintf("No heat map matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
