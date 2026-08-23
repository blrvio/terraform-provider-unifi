package spatialrecord

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

// SpatialRecordDatasourceModel represents the data model for a spatial record data source.
type SpatialRecordDatasourceModel struct {
	base.Model
	Name    types.String `tfsdk:"name"`
	Devices types.List   `tfsdk:"devices"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *SpatialRecordDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *SpatialRecordDatasourceModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.SpatialRecord)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.SpatialRecord, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = ut.StringOrNull(model.Name)

	devices := make([]spatialDeviceModel, 0, len(model.Devices))
	for _, d := range model.Devices {
		pos := spatialPositionModel{
			X: ut.Float64OrNull(d.Position.X),
			Y: ut.Float64OrNull(d.Position.Y),
			Z: ut.Float64OrNull(d.Position.Z),
		}
		posObj, posDiags := types.ObjectValueFrom(ctx, spatialPositionModel{}.AttributeTypes(), pos)
		diags.Append(posDiags...)
		if diags.HasError() {
			return diags
		}
		devices = append(devices, spatialDeviceModel{
			MAC:      ut.StringOrNull(d.MAC),
			Position: posObj,
		})
	}

	list, listDiags := types.ListValueFrom(ctx, spatialDeviceObjectType(), devices)
	diags.Append(listDiags...)
	m.Devices = list
	return diags
}

var (
	_ datasource.DataSource              = &spatialRecordDatasource{}
	_ datasource.DataSourceWithConfigure = &spatialRecordDatasource{}
	_ base.Resource                      = &spatialRecordDatasource{}
)

type spatialRecordDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewSpatialRecordDatasource() datasource.DataSource {
	return &spatialRecordDatasource{}
}

func (d *spatialRecordDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *spatialRecordDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *spatialRecordDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *spatialRecordDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *spatialRecordDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_spatial_record"
}

func (d *spatialRecordDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_spatial_record` data source retrieves an existing spatial record by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the spatial record to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The devices placed within this spatial record.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mac": schema.StringAttribute{
							MarkdownDescription: "The MAC address of the placed device.",
							Computed:            true,
						},
						"position": schema.SingleNestedAttribute{
							MarkdownDescription: "The x/y/z position of the device within the mapped space.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"x": schema.Float64Attribute{MarkdownDescription: "The X coordinate.", Computed: true},
								"y": schema.Float64Attribute{MarkdownDescription: "The Y coordinate.", Computed: true},
								"z": schema.Float64Attribute{MarkdownDescription: "The Z coordinate.", Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

func (d *spatialRecordDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SpatialRecordDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	records, err := d.client.ListSpatialRecord(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list spatial records", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a spatial record.")
		return
	}

	var found *unifi.SpatialRecord
	for i := range records {
		s := records[i]
		if (id != "" && s.ID == id) || (name != "" && s.Name == name) {
			found = &s
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Spatial record not found", fmt.Sprintf("No spatial record matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
