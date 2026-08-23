package unifimap

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := MapModel{
		Lat:        types.StringValue("52.1"),
		Lng:        types.StringValue("-0.5"),
		MapTypeID:  types.StringValue("satellite"),
		Name:       types.StringValue("ground-floor"),
		OffsetLeft: types.Float64Value(12.5),
		OffsetTop:  types.Float64Value(8.25),
		Opacity:    types.Float64Value(0.75),
		Selected:   types.BoolValue(true),
		Tilt:       types.Int32Value(30),
		Type:       types.StringValue("imageMap"),
		Unit:       types.StringValue("m"),
		Upp:        types.Float64Value(1.5),
		Zoom:       types.Int32Value(18),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.Map)
	require.True(t, ok, "Expected model to be *unifi.Map")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "52.1", typed.Lat)
	assert.Equal(t, "-0.5", typed.Lng)
	assert.Equal(t, "satellite", typed.MapTypeID)
	assert.Equal(t, "ground-floor", typed.Name)
	assert.Equal(t, 12.5, typed.OffsetLeft)
	assert.Equal(t, 8.25, typed.OffsetTop)
	assert.Equal(t, 0.75, typed.Opacity)
	assert.Equal(t, true, typed.Selected)
	assert.Equal(t, 30, typed.Tilt)
	assert.Equal(t, "imageMap", typed.Type)
	assert.Equal(t, "m", typed.Unit)
	assert.Equal(t, 1.5, typed.Upp)
	assert.Equal(t, 18, typed.Zoom)
}

func TestMapModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.Map{
		ID:         "merge-id",
		Lat:        "52.1",
		Lng:        "-0.5",
		MapTypeID:  "roadmap",
		Name:       "ground-floor",
		OffsetLeft: 12.5,
		OffsetTop:  8.25,
		Opacity:    0.75,
		Selected:   true,
		Tilt:       30,
		Type:       "googleMap",
		Unit:       "f",
		Upp:        1.5,
		Zoom:       18,
	}

	var d MapModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "52.1", d.Lat.ValueString())
	assert.Equal(t, "-0.5", d.Lng.ValueString())
	assert.Equal(t, "roadmap", d.MapTypeID.ValueString())
	assert.Equal(t, "ground-floor", d.Name.ValueString())
	assert.Equal(t, 12.5, d.OffsetLeft.ValueFloat64())
	assert.Equal(t, 8.25, d.OffsetTop.ValueFloat64())
	assert.Equal(t, 0.75, d.Opacity.ValueFloat64())
	assert.Equal(t, true, d.Selected.ValueBool())
	assert.Equal(t, int32(30), d.Tilt.ValueInt32())
	assert.Equal(t, "googleMap", d.Type.ValueString())
	assert.Equal(t, "f", d.Unit.ValueString())
	assert.Equal(t, 1.5, d.Upp.ValueFloat64())
	assert.Equal(t, int32(18), d.Zoom.ValueInt32())
}

func TestMapModel_Merge_EmptyOptionals(t *testing.T) {
	t.Parallel()

	var d MapModel
	diags := d.Merge(context.Background(), &unifi.Map{ID: "id"})
	assert.False(t, diags.HasError())

	assert.True(t, d.Lat.IsNull())
	assert.True(t, d.Lng.IsNull())
	assert.True(t, d.MapTypeID.IsNull())
	assert.True(t, d.Name.IsNull())
	assert.True(t, d.OffsetLeft.IsNull())
	assert.True(t, d.OffsetTop.IsNull())
	assert.True(t, d.Opacity.IsNull())
	assert.True(t, d.Tilt.IsNull())
	assert.True(t, d.Type.IsNull())
	assert.True(t, d.Unit.IsNull())
	assert.True(t, d.Upp.IsNull())
	assert.True(t, d.Zoom.IsNull())

	// selected is a plain bool (no omitempty), so it always merges to a value.
	assert.False(t, d.Selected.IsNull())
	assert.False(t, d.Selected.ValueBool())
}

func TestMapModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d MapModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
