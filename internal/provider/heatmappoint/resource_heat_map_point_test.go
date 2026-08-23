package heatmappoint

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeatMapPointModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := HeatMapPointModel{
		HeatmapID:     types.StringValue("heatmap-1"),
		DownloadSpeed: types.Float64Value(123.5),
		UploadSpeed:   types.Float64Value(45.25),
		X:             types.Float64Value(10.0),
		Y:             types.Float64Value(20.0),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.HeatMapPoint)
	require.True(t, ok, "Expected model to be *unifi.HeatMapPoint")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "heatmap-1", typed.HeatmapID)
	assert.Equal(t, 123.5, typed.DownloadSpeed)
	assert.Equal(t, 45.25, typed.UploadSpeed)
	assert.Equal(t, 10.0, typed.X)
	assert.Equal(t, 20.0, typed.Y)
}

func TestHeatMapPointModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.HeatMapPoint{
		ID:            "merge-id",
		HeatmapID:     "heatmap-1",
		DownloadSpeed: 500.5,
		UploadSpeed:   250.25,
		X:             1.5,
		Y:             2.5,
	}

	var d HeatMapPointModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "heatmap-1", d.HeatmapID.ValueString())
	assert.Equal(t, 500.5, d.DownloadSpeed.ValueFloat64())
	assert.Equal(t, 250.25, d.UploadSpeed.ValueFloat64())
	assert.Equal(t, 1.5, d.X.ValueFloat64())
	assert.Equal(t, 2.5, d.Y.ValueFloat64())
}

func TestHeatMapPointModel_Merge_EmptyOptionals(t *testing.T) {
	t.Parallel()

	var d HeatMapPointModel
	diags := d.Merge(context.Background(), &unifi.HeatMapPoint{ID: "id", HeatmapID: "heatmap-1"})
	assert.False(t, diags.HasError())
	assert.True(t, d.DownloadSpeed.IsNull())
	assert.True(t, d.UploadSpeed.IsNull())
	assert.True(t, d.X.IsNull())
	assert.True(t, d.Y.IsNull())
}

func TestHeatMapPointModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d HeatMapPointModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
