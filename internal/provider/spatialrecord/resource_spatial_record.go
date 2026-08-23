package spatialrecord

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// spatialPositionModel is the x/y/z position of a device within a spatial record.
type spatialPositionModel struct {
	X types.Float64 `tfsdk:"x"`
	Y types.Float64 `tfsdk:"y"`
	Z types.Float64 `tfsdk:"z"`
}

func (m spatialPositionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"x": types.Float64Type,
		"y": types.Float64Type,
		"z": types.Float64Type,
	}
}

// spatialDeviceModel is one placed device within a spatial record.
type spatialDeviceModel struct {
	MAC      types.String `tfsdk:"mac"`
	Position types.Object `tfsdk:"position"`
}

func (m spatialDeviceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mac":      types.StringType,
		"position": types.ObjectType{AttrTypes: spatialPositionModel{}.AttributeTypes()},
	}
}

func spatialDeviceObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: spatialDeviceModel{}.AttributeTypes()}
}

// SpatialRecordModel represents the data model for a UniFi spatial record.
type SpatialRecordModel struct {
	base.Model
	Name    types.String `tfsdk:"name"`
	Devices types.List   `tfsdk:"devices"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *SpatialRecordModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var devices []spatialDeviceModel
	diags.Append(ut.ListElementsAs(ctx, m.Devices, &devices)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]unifi.SpatialRecordDevices, 0, len(devices))
	for _, d := range devices {
		var pos spatialPositionModel
		if ut.IsDefined(d.Position) {
			diags.Append(d.Position.As(ctx, &pos, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
		}
		out = append(out, unifi.SpatialRecordDevices{
			MAC: d.MAC.ValueString(),
			Position: unifi.SpatialRecordPosition{
				X: pos.X.ValueFloat64(),
				Y: pos.Y.ValueFloat64(),
				Z: pos.Z.ValueFloat64(),
			},
		})
	}

	return &unifi.SpatialRecord{
		ID:      m.ID.ValueString(),
		Name:    m.Name.ValueString(),
		Devices: out,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *SpatialRecordModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.SpatialRecord)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.SpatialRecord, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = ut.StringOrNull(model.Name)

	// Coalesce nil→empty so an empty device list round-trips without a perpetual diff.
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
	_ resource.Resource                = &spatialRecordResource{}
	_ resource.ResourceWithConfigure   = &spatialRecordResource{}
	_ resource.ResourceWithImportState = &spatialRecordResource{}
	_ base.Resource                    = &spatialRecordResource{}
	_ base.ResourceModel               = &SpatialRecordModel{}
)

type spatialRecordResource struct {
	*base.GenericResource[*SpatialRecordModel]
}

// NewSpatialRecordResource creates a new instance of the spatial record resource.
func NewSpatialRecordResource() resource.Resource {
	return &spatialRecordResource{
		GenericResource: base.NewGenericResource(
			"unifi_spatial_record",
			func() *SpatialRecordModel { return &SpatialRecordModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetSpatialRecord(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					spatialRecord, _ := model.(*unifi.SpatialRecord)
					return client.CreateSpatialRecord(ctx, site, spatialRecord)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					spatialRecord, _ := model.(*unifi.SpatialRecord)
					return client.UpdateSpatialRecord(ctx, site, spatialRecord)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteSpatialRecord(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *spatialRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_spatial_record` resource manages a spatial record in the UniFi " +
			"controller: a named set of devices placed at x/y/z coordinates within a mapped space, used " +
			"to describe the physical layout of access points.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the spatial record (1–128 characters).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The devices placed within this spatial record.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mac": schema.StringAttribute{
							MarkdownDescription: "The MAC address of the placed device.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								validators.Mac,
							},
						},
						"position": schema.SingleNestedAttribute{
							MarkdownDescription: "The x/y/z position of the device within the mapped space.",
							Optional:            true,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"x": schema.Float64Attribute{
									MarkdownDescription: "The X coordinate.",
									Optional:            true,
									Computed:            true,
								},
								"y": schema.Float64Attribute{
									MarkdownDescription: "The Y coordinate.",
									Optional:            true,
									Computed:            true,
								},
								"z": schema.Float64Attribute{
									MarkdownDescription: "The Z coordinate.",
									Optional:            true,
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}
