package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperSmtpModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	// Enabled: all server fields reflected.
	model := superSMTPModel{
		Enabled:   types.BoolValue(true),
		Host:      types.StringValue("smtp.example.com"),
		Port:      types.Int32Value(587),
		Sender:    types.StringValue("noreply@example.com"),
		UseAuth:   types.BoolValue(true),
		UseSender: types.BoolValue(true),
		UseSsl:    types.BoolValue(true),
		Username:  types.StringValue("smtp-user"),
		XPassword: types.StringValue("smtp-pass"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperSmtp)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperSmtp")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.Equal(t, "smtp.example.com", typed.Host)
	assert.Equal(t, 587, typed.Port)
	assert.Equal(t, "noreply@example.com", typed.Sender)
	assert.True(t, typed.UseAuth)
	assert.True(t, typed.UseSender)
	assert.True(t, typed.UseSsl)
	assert.Equal(t, "smtp-user", typed.Username)
	assert.Equal(t, "smtp-pass", typed.XPassword)

	// Disabled: server fields are not written.
	disabled := superSMTPModel{
		Enabled:   types.BoolValue(false),
		Host:      types.StringValue("smtp.example.com"),
		Port:      types.Int32Value(587),
		Sender:    types.StringValue("noreply@example.com"),
		UseAuth:   types.BoolValue(true),
		UseSender: types.BoolValue(true),
		UseSsl:    types.BoolValue(true),
		Username:  types.StringValue("smtp-user"),
		XPassword: types.StringValue("smtp-pass"),
	}
	disabled.ID = types.StringValue("test-id")

	unifiModel, diags = disabled.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())
	typed, ok = unifiModel.(*unifi.SettingSuperSmtp)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperSmtp")
	assert.False(t, typed.Enabled)
	assert.Equal(t, "", typed.Host)
	assert.Equal(t, 0, typed.Port)
	assert.Equal(t, "", typed.Sender)
	assert.False(t, typed.UseAuth)
	assert.False(t, typed.UseSender)
	assert.False(t, typed.UseSsl)
	assert.Equal(t, "", typed.Username)
	assert.Equal(t, "", typed.XPassword)
}

func TestSuperSmtpModel_Merge(t *testing.T) {
	t.Parallel()

	// Enabled: all server fields reflected.
	source := &unifi.SettingSuperSmtp{
		ID:        "test-id",
		Enabled:   true,
		Host:      "smtp.example.com",
		Port:      587,
		Sender:    "noreply@example.com",
		UseAuth:   true,
		UseSender: true,
		UseSsl:    true,
		Username:  "smtp-user",
		XPassword: "smtp-pass",
	}

	model := superSMTPModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.True(t, model.Enabled.ValueBool())
	assert.Equal(t, "smtp.example.com", model.Host.ValueString())
	assert.Equal(t, int32(587), model.Port.ValueInt32())
	assert.Equal(t, "noreply@example.com", model.Sender.ValueString())
	assert.True(t, model.UseAuth.ValueBool())
	assert.True(t, model.UseSender.ValueBool())
	assert.True(t, model.UseSsl.ValueBool())
	assert.Equal(t, "smtp-user", model.Username.ValueString())
	assert.Equal(t, "smtp-pass", model.XPassword.ValueString())

	// Disabled: all server fields become null.
	disabled := superSMTPModel{}
	diags = disabled.Merge(context.Background(), &unifi.SettingSuperSmtp{ID: "id", Enabled: false})
	assert.False(t, diags.HasError())

	assert.False(t, disabled.Enabled.ValueBool())
	assert.True(t, disabled.Host.IsNull())
	assert.True(t, disabled.Port.IsNull())
	assert.True(t, disabled.Sender.IsNull())
	assert.True(t, disabled.UseAuth.IsNull())
	assert.True(t, disabled.UseSender.IsNull())
	assert.True(t, disabled.UseSsl.IsNull())
	assert.True(t, disabled.Username.IsNull())
	assert.True(t, disabled.XPassword.IsNull())
}
