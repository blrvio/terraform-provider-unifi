package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestGlobalApModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := globalApModel{
		ApExclusions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("aa:bb:cc:dd:ee:ff"),
		}),
		NaChannelSize:   types.Int32Value(80),
		NaTxPower:       types.Int32Value(20),
		NaTxPowerMode:   types.StringValue("custom"),
		NgChannelSize:   types.Int32Value(20),
		NgTxPower:       types.Int32Value(10),
		NgTxPowerMode:   types.StringValue("auto"),
		SixEChannelSize: types.Int32Value(160),
		SixETxPower:     types.Int32Value(30),
		SixETxPowerMode: types.StringValue("high"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingGlobalAp)
	assert.True(t, ok, "Expected model to be *unifi.SettingGlobalAp")
	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, typed.ApExclusions)
	assert.Equal(t, 80, typed.NaChannelSize)
	assert.Equal(t, 20, typed.NaTxPower)
	assert.Equal(t, "custom", typed.NaTxPowerMode)
	assert.Equal(t, 20, typed.NgChannelSize)
	assert.Equal(t, 10, typed.NgTxPower)
	assert.Equal(t, "auto", typed.NgTxPowerMode)
	assert.Equal(t, 160, typed.SixEChannelSize)
	assert.Equal(t, 30, typed.SixETxPower)
	assert.Equal(t, "high", typed.SixETxPowerMode)
}

func TestGlobalApModel_Merge(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingGlobalAp{
		ID:              "test-id",
		ApExclusions:    []string{"aa:bb:cc:dd:ee:ff"},
		NaChannelSize:   40,
		NaTxPower:       15,
		NaTxPowerMode:   "medium",
		SixEChannelSize: 80,
	}

	var model globalApModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, int32(40), model.NaChannelSize.ValueInt32())
	assert.Equal(t, int32(15), model.NaTxPower.ValueInt32())
	assert.Equal(t, "medium", model.NaTxPowerMode.ValueString())
	assert.Equal(t, int32(80), model.SixEChannelSize.ValueInt32())
	// Zero-valued ints from the API become null in state.
	assert.True(t, model.NgChannelSize.IsNull())
	assert.True(t, model.NgTxPowerMode.IsNull())

	var exclusions []string
	model.ApExclusions.ElementsAs(context.Background(), &exclusions, false)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, exclusions)
}
