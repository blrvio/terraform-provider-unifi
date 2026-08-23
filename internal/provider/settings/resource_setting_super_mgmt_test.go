package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperMgmtModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superMgmtModel{
		AutoUpgrade:                           types.BoolValue(true),
		AutobackupEnabled:                     types.BoolValue(true),
		AutobackupDays:                        types.Int32Value(7),
		AutobackupMaxFiles:                    types.Int32Value(10),
		ContactInfoFullName:                   types.StringValue("Ada Lovelace"),
		ContactInfoCompanyName:                types.StringValue("Analytical Engines"),
		DataRetentionSettingPreference:        types.StringValue("manual"),
		DataRetentionTimeInHoursForDailyScale: types.Int32Value(720),
		LedEnabled:                            types.BoolValue(true),
		LiveChat:                              types.StringValue("super-only"),
		LiveUpdates:                           types.StringValue("auto"),
		StoreEnabled:                          types.StringValue("disabled"),
		GoogleMapsAPIKey:                      types.StringValue("secret-key"),
		XSshUsername:                          types.StringValue("root"),
		XSshPassword:                          types.StringValue("hunter2"),
		AutobackupPostActions: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("copy_local"),
			types.StringValue("copy_cloud"),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperMgmt)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperMgmt")
	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.AutoUpgrade)
	assert.True(t, typed.AutobackupEnabled)
	assert.Equal(t, 7, typed.AutobackupDays)
	assert.Equal(t, 10, typed.AutobackupMaxFiles)
	assert.Equal(t, "Ada Lovelace", typed.ContactInfoFullName)
	assert.Equal(t, "Analytical Engines", typed.ContactInfoCompanyName)
	assert.Equal(t, "manual", typed.DataRetentionSettingPreference)
	assert.Equal(t, 720, typed.DataRetentionTimeInHoursForDailyScale)
	assert.True(t, typed.LedEnabled)
	assert.Equal(t, "super-only", typed.LiveChat)
	assert.Equal(t, "auto", typed.LiveUpdates)
	assert.Equal(t, "disabled", typed.StoreEnabled)
	assert.Equal(t, "secret-key", typed.GoogleMapsApiKey)
	assert.Equal(t, "root", typed.XSshUsername)
	assert.Equal(t, "hunter2", typed.XSshPassword)
	assert.Equal(t, []string{"copy_local", "copy_cloud"}, typed.AutobackupPostActions)
}

func TestSuperMgmtModel_Merge(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingSuperMgmt{
		ID:                             "test-id",
		AutoUpgrade:                    true,
		AutobackupEnabled:              true,
		AutobackupDays:                 7,
		ContactInfoFullName:            "Ada Lovelace",
		DataRetentionSettingPreference: "auto",
		LedEnabled:                     true,
		GoogleMapsApiKey:               "secret-key",
		AutobackupPostActions:          []string{"copy_local"},
	}

	var model superMgmtModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.True(t, model.AutoUpgrade.ValueBool())
	assert.True(t, model.AutobackupEnabled.ValueBool())
	assert.Equal(t, int32(7), model.AutobackupDays.ValueInt32())
	assert.Equal(t, "Ada Lovelace", model.ContactInfoFullName.ValueString())
	assert.Equal(t, "auto", model.DataRetentionSettingPreference.ValueString())
	assert.True(t, model.LedEnabled.ValueBool())
	assert.Equal(t, "secret-key", model.GoogleMapsAPIKey.ValueString())
	// Empty string fields from the API become null in state.
	assert.True(t, model.ContactInfoCompanyName.IsNull())
	assert.True(t, model.AutobackupMaxFiles.IsNull())

	var postActions []string
	model.AutobackupPostActions.ElementsAs(context.Background(), &postActions, false)
	assert.Equal(t, []string{"copy_local"}, postActions)
}

func TestSuperMgmtModel_Merge_EmptyPostActions(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingSuperMgmt{ID: "test-id"}

	var model superMgmtModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.False(t, model.AutobackupPostActions.IsNull())
	assert.Equal(t, 0, len(model.AutobackupPostActions.Elements()))
}
