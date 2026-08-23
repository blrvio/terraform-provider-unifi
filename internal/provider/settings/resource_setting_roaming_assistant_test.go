package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestRoamingAssistantModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := roamingAssistantModel{
		Enabled: types.BoolValue(true),
		Rssi:    types.Int32Value(-70),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingRoamingAssistant)
	assert.True(t, ok, "Expected model to be *unifi.SettingRoamingAssistant")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.Equal(t, -70, typed.Rssi)
}

func TestRoamingAssistantModel_AsUnifiModel_Disabled(t *testing.T) {
	t.Parallel()

	model := roamingAssistantModel{
		Enabled: types.BoolValue(false),
		Rssi:    types.Int32Value(-70),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingRoamingAssistant)
	assert.True(t, ok)

	assert.False(t, typed.Enabled)
	// When disabled, dependent fields are left at their zero value.
	assert.Equal(t, 0, typed.Rssi)
}

func TestRoamingAssistantModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingRoamingAssistant{
		ID:      "merge-id",
		Enabled: true,
		Rssi:    -65,
	}

	var d roamingAssistantModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.Enabled.ValueBool())
	assert.Equal(t, int32(-65), d.Rssi.ValueInt32())
}

func TestRoamingAssistantModel_Merge_Disabled(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingRoamingAssistant{
		ID:      "merge-id",
		Enabled: false,
		Rssi:    -65,
	}

	var d roamingAssistantModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.False(t, d.Enabled.ValueBool())
	// When disabled, dependent fields become null.
	assert.True(t, d.Rssi.IsNull())
}
