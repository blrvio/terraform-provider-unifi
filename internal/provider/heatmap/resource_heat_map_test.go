package heatmap

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeatMapModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := HeatMapModel{
		MapID:       types.StringValue("map-1"),
		Name:        types.StringValue("ground-floor"),
		Description: types.StringValue("coverage"),
		Type:        types.StringValue("download"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.HeatMap)
	require.True(t, ok, "Expected model to be *unifi.HeatMap")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "map-1", typed.MapID)
	assert.Equal(t, "ground-floor", typed.Name)
	assert.Equal(t, "coverage", typed.Description)
	assert.Equal(t, "download", typed.Type)
}

func TestHeatMapModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.HeatMap{
		ID:          "merge-id",
		MapID:       "map-1",
		Name:        "ground-floor",
		Description: "coverage",
		Type:        "upload",
	}

	var d HeatMapModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "map-1", d.MapID.ValueString())
	assert.Equal(t, "ground-floor", d.Name.ValueString())
	assert.Equal(t, "coverage", d.Description.ValueString())
	assert.Equal(t, "upload", d.Type.ValueString())
}

func TestHeatMapModel_Merge_EmptyOptionals(t *testing.T) {
	t.Parallel()

	var d HeatMapModel
	diags := d.Merge(context.Background(), &unifi.HeatMap{ID: "id", MapID: "map-1"})
	assert.False(t, diags.HasError())
	assert.True(t, d.Name.IsNull())
	assert.True(t, d.Description.IsNull())
	assert.True(t, d.Type.IsNull())
}

func TestHeatMapModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d HeatMapModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
