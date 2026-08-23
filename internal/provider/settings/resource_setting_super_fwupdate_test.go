package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperFwupdateModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superFwupdateModel{
		ControllerChannel: types.StringValue("release"),
		FirmwareChannel:   types.StringValue("beta"),
		SsoEnabled:        types.BoolValue(true),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperFwupdate)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperFwupdate")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "release", typed.ControllerChannel)
	assert.Equal(t, "beta", typed.FirmwareChannel)
	assert.True(t, typed.SsoEnabled)
}

func TestSuperFwupdateModel_Merge(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSuperFwupdate{
		ID:                "test-id",
		ControllerChannel: "release",
		FirmwareChannel:   "beta",
		SsoEnabled:        true,
	}

	model := superFwupdateModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "release", model.ControllerChannel.ValueString())
	assert.Equal(t, "beta", model.FirmwareChannel.ValueString())
	assert.True(t, model.SsoEnabled.ValueBool())

	// Empty channels round-trip to null; bool always known.
	empty := superFwupdateModel{}
	diags = empty.Merge(context.Background(), &unifi.SettingSuperFwupdate{ID: "id"})
	assert.False(t, diags.HasError())
	assert.True(t, empty.ControllerChannel.IsNull())
	assert.True(t, empty.FirmwareChannel.IsNull())
	assert.False(t, empty.SsoEnabled.IsNull())
	assert.False(t, empty.SsoEnabled.ValueBool())
}
