package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestBaresipModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := baresipModel{
		Enabled:       types.BoolValue(true),
		OutboundProxy: types.StringValue("proxy.example.com"),
		PackageURL:    types.StringValue("https://example.com/baresip.pkg"),
		Server:        types.StringValue("sip.example.com"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingBaresip)
	assert.True(t, ok, "Expected model to be *unifi.SettingBaresip")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Enabled)
	assert.Equal(t, "proxy.example.com", typed.OutboundProxy)
	assert.Equal(t, "https://example.com/baresip.pkg", typed.PackageUrl)
	assert.Equal(t, "sip.example.com", typed.Server)
}

func TestBaresipModel_AsUnifiModel_Disabled(t *testing.T) {
	t.Parallel()

	model := baresipModel{
		Enabled:       types.BoolValue(false),
		OutboundProxy: types.StringValue("proxy.example.com"),
		PackageURL:    types.StringValue("https://example.com/baresip.pkg"),
		Server:        types.StringValue("sip.example.com"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingBaresip)
	assert.True(t, ok)

	assert.False(t, typed.Enabled)
	assert.Equal(t, "", typed.OutboundProxy)
	assert.Equal(t, "", typed.PackageUrl)
	assert.Equal(t, "", typed.Server)
}

func TestBaresipModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingBaresip{
		ID:            "merge-id",
		Enabled:       true,
		OutboundProxy: "proxy.example.com",
		PackageUrl:    "https://example.com/baresip.pkg",
		Server:        "sip.example.com",
	}

	var d baresipModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.Enabled.ValueBool())
	assert.Equal(t, "proxy.example.com", d.OutboundProxy.ValueString())
	assert.Equal(t, "https://example.com/baresip.pkg", d.PackageURL.ValueString())
	assert.Equal(t, "sip.example.com", d.Server.ValueString())
}

func TestBaresipModel_Merge_Disabled(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingBaresip{
		ID:            "merge-id",
		Enabled:       false,
		OutboundProxy: "proxy.example.com",
		PackageUrl:    "https://example.com/baresip.pkg",
		Server:        "sip.example.com",
	}

	var d baresipModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.False(t, d.Enabled.ValueBool())
	assert.True(t, d.OutboundProxy.IsNull())
	assert.True(t, d.PackageURL.IsNull())
	assert.True(t, d.Server.IsNull())
}
