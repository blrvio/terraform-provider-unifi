package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSnmpModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := snmpModel{
		Enabled:   types.BoolValue(true),
		EnabledV3: types.BoolValue(false),
		Community: types.StringValue("public"),
		Username:  types.StringValue("monitor"),
		XPassword: types.StringValue("s3cr3tpass"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSnmp)
	assert.True(t, ok, "Expected model to be *unifi.SettingSnmp")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.False(t, typed.EnabledV3)
	assert.Equal(t, "public", typed.Community)
	assert.Equal(t, "monitor", typed.Username)
	assert.Equal(t, "s3cr3tpass", typed.XPassword)
}

func TestSnmpModel_AsUnifiModel_EmptyOptionals(t *testing.T) {
	t.Parallel()

	model := snmpModel{
		Enabled:   types.BoolValue(false),
		EnabledV3: types.BoolValue(false),
		Community: types.StringNull(),
		Username:  types.StringValue(""),
		XPassword: types.StringUnknown(),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSnmp)
	assert.True(t, ok, "Expected model to be *unifi.SettingSnmp")

	// Null/empty/unknown optionals normalize to the zero string.
	assert.Equal(t, "", typed.Community)
	assert.Equal(t, "", typed.Username)
	assert.Equal(t, "", typed.XPassword)
}

func TestSnmpModel_Merge(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSnmp{
		ID:        "merged-id",
		Enabled:   true,
		EnabledV3: true,
		Community: "private",
		Username:  "admin",
		XPassword: "anotherpass",
	}

	var model snmpModel
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merged-id", model.ID.ValueString())
	assert.True(t, model.Enabled.ValueBool())
	assert.True(t, model.EnabledV3.ValueBool())
	assert.Equal(t, "private", model.Community.ValueString())
	assert.Equal(t, "admin", model.Username.ValueString())
	assert.Equal(t, "anotherpass", model.XPassword.ValueString())
}

func TestSnmpModel_Merge_EmptyOptionalsBecomeNull(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSnmp{
		ID:      "merged-id",
		Enabled: false,
	}

	var model snmpModel
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.True(t, model.Community.IsNull())
	assert.True(t, model.Username.IsNull())
	assert.True(t, model.XPassword.IsNull())
}

func TestSnmpModel_Merge_WrongType(t *testing.T) {
	t.Parallel()

	var model snmpModel
	diags := model.Merge(context.Background(), &unifi.SettingNtp{})
	assert.True(t, diags.HasError())
}
