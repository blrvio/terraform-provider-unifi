package virtualdevice

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualDeviceModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := VirtualDeviceModel{
		MapID:          types.StringValue("map-1"),
		Type:           types.StringValue("uap"),
		X:              types.StringValue("120.5"),
		Y:              types.StringValue("340.0"),
		HeightInMeters: types.Float64Value(2.7),
		Locked:         types.BoolValue(true),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.VirtualDevice)
	require.True(t, ok, "Expected model to be *unifi.VirtualDevice")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "map-1", typed.MapID)
	assert.Equal(t, "uap", typed.Type)
	assert.Equal(t, "120.5", typed.X)
	assert.Equal(t, "340.0", typed.Y)
	assert.InDelta(t, 2.7, typed.HeightInMeters, 0.0001)
	assert.True(t, typed.Locked)
}

func TestVirtualDeviceModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.VirtualDevice{
		ID:             "merge-id",
		MapID:          "map-1",
		Type:           "usw",
		X:              "10",
		Y:              "20",
		HeightInMeters: 3.0,
		Locked:         false,
	}

	var d VirtualDeviceModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "map-1", d.MapID.ValueString())
	assert.Equal(t, "usw", d.Type.ValueString())
	assert.Equal(t, "10", d.X.ValueString())
	assert.Equal(t, "20", d.Y.ValueString())
	assert.InDelta(t, 3.0, d.HeightInMeters.ValueFloat64(), 0.0001)
	assert.False(t, d.Locked.ValueBool())
}

func TestVirtualDeviceModel_Merge_ZeroHeightIsNull(t *testing.T) {
	t.Parallel()

	var d VirtualDeviceModel
	diags := d.Merge(context.Background(), &unifi.VirtualDevice{ID: "id", MapID: "m", Type: "uap", X: "0", Y: "0"})
	assert.False(t, diags.HasError())
	assert.True(t, d.HeightInMeters.IsNull(), "a zero height round-trips to null")
}

func TestVirtualDeviceModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d VirtualDeviceModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
