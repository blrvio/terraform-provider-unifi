package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMdnsModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	customServices, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&mdnsCustomServiceModel{}).AttributeTypes()}, []mdnsCustomServiceModel{
		{
			Address: types.StringValue("_airplay._tcp.local"),
			Name:    types.StringValue("AirPlay"),
		},
	})
	assert.False(t, diags.HasError())

	predefinedServices, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&mdnsPredefinedServiceModel{}).AttributeTypes()}, []mdnsPredefinedServiceModel{
		{Code: types.StringValue("google_chromecast")},
	})
	assert.False(t, diags.HasError())

	model := mdnsModel{
		Mode:               types.StringValue("custom"),
		CustomServices:     customServices,
		PredefinedServices: predefinedServices,
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(ctx)
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingMdns)
	assert.True(t, ok, "Expected model to be *unifi.SettingMdns")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "custom", typed.Mode)

	assert.Len(t, typed.CustomServices, 1)
	assert.Equal(t, "_airplay._tcp.local", typed.CustomServices[0].Address)
	assert.Equal(t, "AirPlay", typed.CustomServices[0].Name)

	assert.Len(t, typed.PredefinedServices, 1)
	assert.Equal(t, "google_chromecast", typed.PredefinedServices[0].Code)
}

func TestMdnsModel_Merge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	source := &unifi.SettingMdns{
		ID:   "test-id",
		Mode: "custom",
		CustomServices: []unifi.SettingMdnsCustomServices{
			{Address: "_ipp._tcp.local", Name: "Printers"},
		},
		PredefinedServices: []unifi.SettingMdnsPredefinedServices{
			{Code: "sonos"},
		},
	}

	var model mdnsModel
	diags := model.Merge(ctx, source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "custom", model.Mode.ValueString())

	var customServices []mdnsCustomServiceModel
	diags = model.CustomServices.ElementsAs(ctx, &customServices, false)
	assert.False(t, diags.HasError())
	assert.Len(t, customServices, 1)
	assert.Equal(t, "_ipp._tcp.local", customServices[0].Address.ValueString())
	assert.Equal(t, "Printers", customServices[0].Name.ValueString())

	var predefinedServices []mdnsPredefinedServiceModel
	diags = model.PredefinedServices.ElementsAs(ctx, &predefinedServices, false)
	assert.False(t, diags.HasError())
	assert.Len(t, predefinedServices, 1)
	assert.Equal(t, "sonos", predefinedServices[0].Code.ValueString())
}
