package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNetflowModel_AsUnifiModel_Enabled(t *testing.T) {
	t.Parallel()

	model := netflowModel{
		Enabled:             types.BoolValue(true),
		AutoEngineIDEnabled: types.BoolValue(true),
		EngineID:            types.Int32Value(1),
		ExportFrequency:     types.Int32Value(60),
		NetworkIDs: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("net-1"),
		}),
		Port:         types.Int32Value(2055),
		RefreshRate:  types.Int32Value(30),
		SamplingMode: types.StringValue("hash"),
		SamplingRate: types.Int32Value(100),
		Server:       types.StringValue("10.0.0.5"),
		Version:      types.Int32Value(9),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingNetflow)
	assert.True(t, ok, "Expected model to be *unifi.SettingNetflow")
	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.True(t, typed.AutoEngineIDEnabled)
	assert.Equal(t, 1, typed.EngineID)
	assert.Equal(t, 60, typed.ExportFrequency)
	assert.Equal(t, []string{"net-1"}, typed.NetworkIDs)
	assert.Equal(t, 2055, typed.Port)
	assert.Equal(t, 30, typed.RefreshRate)
	assert.Equal(t, "hash", typed.SamplingMode)
	assert.Equal(t, 100, typed.SamplingRate)
	assert.Equal(t, "10.0.0.5", typed.Server)
	assert.Equal(t, 9, typed.Version)
}

func TestNetflowModel_AsUnifiModel_Disabled(t *testing.T) {
	t.Parallel()

	// When disabled, optional fields must not be propagated even if set.
	model := netflowModel{
		Enabled: types.BoolValue(false),
		Port:    types.Int32Value(2055),
		Server:  types.StringValue("10.0.0.5"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingNetflow)
	assert.True(t, ok, "Expected model to be *unifi.SettingNetflow")
	assert.False(t, typed.Enabled)
	assert.Equal(t, 0, typed.Port)
	assert.Equal(t, "", typed.Server)
	assert.Equal(t, []string{}, typed.NetworkIDs)
}

func TestNetflowModel_Merge_Enabled(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingNetflow{
		ID:           "test-id",
		Enabled:      true,
		Port:         2055,
		SamplingMode: "random",
		Version:      10,
		NetworkIDs:   []string{"net-1", "net-2"},
	}

	var model netflowModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.True(t, model.Enabled.ValueBool())
	assert.Equal(t, int32(2055), model.Port.ValueInt32())
	assert.Equal(t, "random", model.SamplingMode.ValueString())
	assert.Equal(t, int32(10), model.Version.ValueInt32())

	var networkIDs []string
	model.NetworkIDs.ElementsAs(context.Background(), &networkIDs, false)
	assert.Equal(t, []string{"net-1", "net-2"}, networkIDs)
}

func TestNetflowModel_Merge_Disabled(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingNetflow{ID: "test-id", Enabled: false}

	var model netflowModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.False(t, model.Enabled.ValueBool())
	assert.True(t, model.AutoEngineIDEnabled.IsNull())
	assert.True(t, model.Port.IsNull())
	assert.True(t, model.SamplingMode.IsNull())
	assert.True(t, model.Server.IsNull())
	assert.True(t, model.Version.IsNull())
	assert.False(t, model.NetworkIDs.IsNull())
	assert.Equal(t, 0, len(model.NetworkIDs.Elements()))
}
