package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperCloudaccessModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	// Enabled: credential fields are reflected; device_id/ubic_uuid always set.
	model := superCloudaccessModel{
		Enabled:         types.BoolValue(true),
		DeviceID:        types.StringValue("dev-123"),
		UbicUUID:        types.StringValue("ubic-uuid"),
		DeviceAuth:      types.StringValue("auth-secret"),
		XCertificateArn: types.StringValue("arn:aws:iot:cert"),
		XCertificatePem: types.StringValue("-----BEGIN CERT-----"),
		XPrivateKey:     types.StringValue("-----BEGIN KEY-----"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperCloudaccess)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperCloudaccess")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.Equal(t, "dev-123", typed.DeviceID)
	assert.Equal(t, "ubic-uuid", typed.UbicUuid)
	assert.Equal(t, "auth-secret", typed.DeviceAuth)
	assert.Equal(t, "arn:aws:iot:cert", typed.XCertificateArn)
	assert.Equal(t, "-----BEGIN CERT-----", typed.XCertificatePem)
	assert.Equal(t, "-----BEGIN KEY-----", typed.XPrivateKey)

	// Disabled: credential fields are not written, but device_id/ubic_uuid still are.
	disabled := superCloudaccessModel{
		Enabled:         types.BoolValue(false),
		DeviceID:        types.StringValue("dev-123"),
		UbicUUID:        types.StringValue("ubic-uuid"),
		DeviceAuth:      types.StringValue("auth-secret"),
		XCertificateArn: types.StringValue("arn:aws:iot:cert"),
		XCertificatePem: types.StringValue("-----BEGIN CERT-----"),
		XPrivateKey:     types.StringValue("-----BEGIN KEY-----"),
	}
	disabled.ID = types.StringValue("test-id")

	unifiModel, diags = disabled.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())
	typed, ok = unifiModel.(*unifi.SettingSuperCloudaccess)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperCloudaccess")
	assert.False(t, typed.Enabled)
	assert.Equal(t, "dev-123", typed.DeviceID)
	assert.Equal(t, "ubic-uuid", typed.UbicUuid)
	assert.Equal(t, "", typed.DeviceAuth)
	assert.Equal(t, "", typed.XCertificateArn)
	assert.Equal(t, "", typed.XCertificatePem)
	assert.Equal(t, "", typed.XPrivateKey)
}

func TestSuperCloudaccessModel_Merge(t *testing.T) {
	t.Parallel()

	// Enabled: credential fields reflected.
	source := &unifi.SettingSuperCloudaccess{
		ID:              "test-id",
		Enabled:         true,
		DeviceID:        "dev-123",
		UbicUuid:        "ubic-uuid",
		DeviceAuth:      "auth-secret",
		XCertificateArn: "arn:aws:iot:cert",
		XCertificatePem: "-----BEGIN CERT-----",
		XPrivateKey:     "-----BEGIN KEY-----",
	}

	model := superCloudaccessModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.True(t, model.Enabled.ValueBool())
	assert.Equal(t, "dev-123", model.DeviceID.ValueString())
	assert.Equal(t, "ubic-uuid", model.UbicUUID.ValueString())
	assert.Equal(t, "auth-secret", model.DeviceAuth.ValueString())
	assert.Equal(t, "arn:aws:iot:cert", model.XCertificateArn.ValueString())
	assert.Equal(t, "-----BEGIN CERT-----", model.XCertificatePem.ValueString())
	assert.Equal(t, "-----BEGIN KEY-----", model.XPrivateKey.ValueString())

	// Disabled: credential fields become null; device_id/ubic_uuid still reflected.
	disabledSource := &unifi.SettingSuperCloudaccess{
		ID:       "test-id",
		Enabled:  false,
		DeviceID: "dev-123",
		UbicUuid: "ubic-uuid",
	}

	disabled := superCloudaccessModel{}
	diags = disabled.Merge(context.Background(), disabledSource)
	assert.False(t, diags.HasError())

	assert.False(t, disabled.Enabled.ValueBool())
	assert.Equal(t, "dev-123", disabled.DeviceID.ValueString())
	assert.Equal(t, "ubic-uuid", disabled.UbicUUID.ValueString())
	assert.True(t, disabled.DeviceAuth.IsNull())
	assert.True(t, disabled.XCertificateArn.IsNull())
	assert.True(t, disabled.XCertificatePem.IsNull())
	assert.True(t, disabled.XPrivateKey.IsNull())
}
