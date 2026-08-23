package spatialrecord

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deviceList(t *testing.T, devices ...spatialDeviceModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), spatialDeviceObjectType(), devices)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

func positionObject(t *testing.T, x, y, z float64) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), spatialPositionModel{}.AttributeTypes(), spatialPositionModel{
		X: types.Float64Value(x),
		Y: types.Float64Value(y),
		Z: types.Float64Value(z),
	})
	require.False(t, diags.HasError(), "ObjectValueFrom diagnostics: %v", diags)
	return obj
}

func TestSpatialRecordModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := SpatialRecordModel{
		Name: types.StringValue("floor-1"),
		Devices: deviceList(t, spatialDeviceModel{
			MAC:      types.StringValue("aa:bb:cc:dd:ee:ff"),
			Position: positionObject(t, 1.5, 2.5, 3.5),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SpatialRecord)
	require.True(t, ok, "Expected model to be *unifi.SpatialRecord")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "floor-1", typed.Name)
	require.Len(t, typed.Devices, 1)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", typed.Devices[0].MAC)
	assert.Equal(t, 1.5, typed.Devices[0].Position.X)
	assert.Equal(t, 2.5, typed.Devices[0].Position.Y)
	assert.Equal(t, 3.5, typed.Devices[0].Position.Z)
}

func TestSpatialRecordModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SpatialRecord{
		ID:   "merge-id",
		Name: "floor-1",
		Devices: []unifi.SpatialRecordDevices{
			{
				MAC:      "aa:bb:cc:dd:ee:ff",
				Position: unifi.SpatialRecordPosition{X: 1.5, Y: 2.5, Z: 3.5},
			},
		},
	}

	var d SpatialRecordModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "floor-1", d.Name.ValueString())

	var devices []spatialDeviceModel
	d.Devices.ElementsAs(context.Background(), &devices, false)
	require.Len(t, devices, 1)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", devices[0].MAC.ValueString())
}

func TestSpatialRecordModel_Merge_EmptyDevices(t *testing.T) {
	t.Parallel()

	var d SpatialRecordModel
	diags := d.Merge(context.Background(), &unifi.SpatialRecord{ID: "id", Name: "empty"})
	assert.False(t, diags.HasError())
	assert.False(t, d.Devices.IsNull())
	assert.Equal(t, 0, len(d.Devices.Elements()))
}

func TestSpatialRecordModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d SpatialRecordModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
