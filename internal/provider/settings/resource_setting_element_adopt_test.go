package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestElementAdoptModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := elementAdoptModel{
		Enabled:       types.BoolValue(true),
		XElementEssid: types.StringValue("element-ssid"),
		XElementPsk:   types.StringValue("super-secret"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingElementAdopt)
	assert.True(t, ok, "Expected model to be *unifi.SettingElementAdopt")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.Equal(t, "element-ssid", typed.XElementEssid)
	assert.Equal(t, "super-secret", typed.XElementPsk)
}

func TestElementAdoptModel_AsUnifiModel_Disabled(t *testing.T) {
	t.Parallel()

	model := elementAdoptModel{
		Enabled:       types.BoolValue(false),
		XElementEssid: types.StringValue("element-ssid"),
		XElementPsk:   types.StringValue("super-secret"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingElementAdopt)
	assert.True(t, ok)

	assert.False(t, typed.Enabled)
	assert.Equal(t, "", typed.XElementEssid)
	assert.Equal(t, "", typed.XElementPsk)
}

func TestElementAdoptModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingElementAdopt{
		ID:            "merge-id",
		Enabled:       true,
		XElementEssid: "element-ssid",
		XElementPsk:   "super-secret",
	}

	var d elementAdoptModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.Enabled.ValueBool())
	assert.Equal(t, "element-ssid", d.XElementEssid.ValueString())
	assert.Equal(t, "super-secret", d.XElementPsk.ValueString())
}

func TestElementAdoptModel_Merge_Disabled(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingElementAdopt{
		ID:            "merge-id",
		Enabled:       false,
		XElementEssid: "element-ssid",
		XElementPsk:   "super-secret",
	}

	var d elementAdoptModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.False(t, d.Enabled.ValueBool())
	assert.True(t, d.XElementEssid.IsNull())
	assert.True(t, d.XElementPsk.IsNull())
}
