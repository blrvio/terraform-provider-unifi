package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTrafficFlowModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := trafficFlowModel{
		EnabledAllowedTraffic:        types.BoolValue(true),
		GatewayDNSEnabled:            types.BoolValue(false),
		UnifiDeviceManagementEnabled: types.BoolValue(true),
		UnifiServicesEnabled:         types.BoolValue(false),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingTrafficFlow)
	assert.True(t, ok, "Expected model to be *unifi.SettingTrafficFlow")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.EnabledAllowedTraffic)
	assert.False(t, typed.GatewayDNSEnabled)
	assert.True(t, typed.UnifiDeviceManagementEnabled)
	assert.False(t, typed.UnifiServicesEnabled)
}

func TestTrafficFlowModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingTrafficFlow{
		ID:                           "merge-id",
		EnabledAllowedTraffic:        true,
		GatewayDNSEnabled:            true,
		UnifiDeviceManagementEnabled: false,
		UnifiServicesEnabled:         true,
	}

	var d trafficFlowModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.EnabledAllowedTraffic.ValueBool())
	assert.True(t, d.GatewayDNSEnabled.ValueBool())
	assert.False(t, d.UnifiDeviceManagementEnabled.ValueBool())
	assert.True(t, d.UnifiServicesEnabled.ValueBool())
}
