package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestBroadcastModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := broadcastModel{
		SoundAfterEnabled:   types.BoolValue(true),
		SoundAfterResource:  types.StringValue("res-after"),
		SoundAfterType:      types.StringValue("sample"),
		SoundBeforeEnabled:  types.BoolValue(false),
		SoundBeforeResource: types.StringValue("res-before"),
		SoundBeforeType:     types.StringValue("media"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingBroadcast)
	assert.True(t, ok, "Expected model to be *unifi.SettingBroadcast")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.SoundAfterEnabled)
	assert.Equal(t, "res-after", typed.SoundAfterResource)
	assert.Equal(t, "sample", typed.SoundAfterType)
	assert.False(t, typed.SoundBeforeEnabled)
	assert.Equal(t, "res-before", typed.SoundBeforeResource)
	assert.Equal(t, "media", typed.SoundBeforeType)
}

func TestBroadcastModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingBroadcast{
		ID:                  "merge-id",
		SoundAfterEnabled:   true,
		SoundAfterResource:  "res-after",
		SoundAfterType:      "media",
		SoundBeforeEnabled:  false,
		SoundBeforeResource: "",
		SoundBeforeType:     "",
	}

	var d broadcastModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.SoundAfterEnabled.ValueBool())
	assert.Equal(t, "res-after", d.SoundAfterResource.ValueString())
	assert.Equal(t, "media", d.SoundAfterType.ValueString())
	assert.False(t, d.SoundBeforeEnabled.ValueBool())
	// Empty strings round-trip to null.
	assert.True(t, d.SoundBeforeResource.IsNull())
	assert.True(t, d.SoundBeforeType.IsNull())
}
