package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperSdnModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superSdnModel{
		AuthToken:       types.StringValue("secret-token"),
		DeviceID:        types.StringValue("dev-123"),
		Enabled:         types.BoolValue(true),
		Migrated:        types.BoolValue(true),
		SsoLoginEnabled: types.StringValue("true"),
		UbicUUID:        types.StringValue("ubic-uuid"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperSdn)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperSdn")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "secret-token", typed.AuthToken)
	assert.Equal(t, "dev-123", typed.DeviceID)
	assert.True(t, typed.Enabled)
	assert.True(t, typed.Migrated)
	assert.Equal(t, "true", typed.SsoLoginEnabled)
	assert.Equal(t, "ubic-uuid", typed.UbicUuid)
}

func TestSuperSdnModel_Merge(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSuperSdn{
		ID:              "test-id",
		AuthToken:       "secret-token",
		DeviceID:        "dev-123",
		Enabled:         true,
		Migrated:        true,
		SsoLoginEnabled: "true",
		UbicUuid:        "ubic-uuid",
	}

	model := superSdnModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "secret-token", model.AuthToken.ValueString())
	assert.Equal(t, "dev-123", model.DeviceID.ValueString())
	assert.True(t, model.Enabled.ValueBool())
	assert.True(t, model.Migrated.ValueBool())
	assert.Equal(t, "true", model.SsoLoginEnabled.ValueString())
	assert.Equal(t, "ubic-uuid", model.UbicUUID.ValueString())

	// Empty strings round-trip to null; bools always known.
	empty := superSdnModel{}
	diags = empty.Merge(context.Background(), &unifi.SettingSuperSdn{ID: "id"})
	assert.False(t, diags.HasError())
	assert.True(t, empty.AuthToken.IsNull())
	assert.True(t, empty.DeviceID.IsNull())
	assert.True(t, empty.SsoLoginEnabled.IsNull())
	assert.True(t, empty.UbicUUID.IsNull())
	assert.False(t, empty.Enabled.IsNull())
	assert.False(t, empty.Migrated.IsNull())
}
