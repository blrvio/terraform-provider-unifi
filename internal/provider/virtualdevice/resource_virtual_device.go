package virtualdevice

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// VirtualDeviceModel represents the data model for a UniFi virtual device (a
// device icon placed on a topology / floor-plan map).
type VirtualDeviceModel struct {
	base.Model
	MapID          types.String  `tfsdk:"map_id"`
	Type           types.String  `tfsdk:"type"`
	X              types.String  `tfsdk:"x"`
	Y              types.String  `tfsdk:"y"`
	HeightInMeters types.Float64 `tfsdk:"height_in_meters"`
	Locked         types.Bool    `tfsdk:"locked"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *VirtualDeviceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.VirtualDevice{
		ID:             m.ID.ValueString(),
		MapID:          m.MapID.ValueString(),
		Type:           m.Type.ValueString(),
		X:              m.X.ValueString(),
		Y:              m.Y.ValueString(),
		HeightInMeters: m.HeightInMeters.ValueFloat64(),
		Locked:         m.Locked.ValueBool(),
	}, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *VirtualDeviceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.VirtualDevice)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.VirtualDevice, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.MapID = types.StringValue(model.MapID)
	m.Type = types.StringValue(model.Type)
	m.X = types.StringValue(model.X)
	m.Y = types.StringValue(model.Y)
	// height_in_meters is omitempty; a zero round-trips to null so an unset
	// height does not produce a perpetual diff.
	if model.HeightInMeters == 0 {
		m.HeightInMeters = types.Float64Null()
	} else {
		m.HeightInMeters = types.Float64Value(model.HeightInMeters)
	}
	m.Locked = types.BoolValue(model.Locked)
	return diag.Diagnostics{}
}

var (
	_ resource.Resource                = &virtualDeviceResource{}
	_ resource.ResourceWithConfigure   = &virtualDeviceResource{}
	_ resource.ResourceWithImportState = &virtualDeviceResource{}
	_ base.Resource                    = &virtualDeviceResource{}
	_ base.ResourceModel               = &VirtualDeviceModel{}
)

type virtualDeviceResource struct {
	*base.GenericResource[*VirtualDeviceModel]
}

// NewVirtualDeviceResource creates a new instance of the virtual device resource.
func NewVirtualDeviceResource() resource.Resource {
	return &virtualDeviceResource{
		GenericResource: base.NewGenericResource(
			"unifi_virtual_device",
			func() *VirtualDeviceModel { return &VirtualDeviceModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetVirtualDevice(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					vd, _ := model.(*unifi.VirtualDevice)
					return client.CreateVirtualDevice(ctx, site, vd)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					vd, _ := model.(*unifi.VirtualDevice)
					return client.UpdateVirtualDevice(ctx, site, vd)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteVirtualDevice(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *virtualDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_virtual_device` resource places a device icon on a UniFi topology or " +
			"floor-plan map.\n\n" +
			"A virtual device is a **map-placement artifact**, not a network-configuration object: it records " +
			"the position (`x`, `y`, `height_in_meters`) and `type` of a device icon on a map (`map_id`). Use " +
			"it to lay out planned or physical devices on a floor plan; it does not configure any live device.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"map_id": schema.StringAttribute{
				MarkdownDescription: "The id of the map on which the device is placed.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The device type represented by the icon. One of `uap`, `usg`, `usw`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("uap", "usg", "usw"),
				},
			},
			"x": schema.StringAttribute{
				MarkdownDescription: "The horizontal position of the icon on the map.",
				Required:            true,
			},
			"y": schema.StringAttribute{
				MarkdownDescription: "The vertical position of the icon on the map.",
				Required:            true,
			},
			"height_in_meters": schema.Float64Attribute{
				MarkdownDescription: "The mounting height of the device in meters.",
				Optional:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the icon is locked in place on the map. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}
