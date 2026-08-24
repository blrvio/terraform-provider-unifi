package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsgGeoIsNoOp verifies that Region Blocking is DEPRECATED and a no-op on
// unifi_setting_usg: even when geo_ip_filtering is set in config, AsUnifiModel
// must NOT enable/write it (the controller does not persist these flat fields on
// UDM-Pro / Network 10.x — use unifi_setting_usg_geo). This prevents the
// "inconsistent result after apply" that forced ignore_changes.
func TestUsgGeoIsNoOp(t *testing.T) {
	ctx := context.Background()

	geo := &GeoIPFilteringModel{
		Mode:             types.StringValue("block"),
		TrafficDirection: types.StringValue("both"),
		Countries:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("RU")}),
	}
	geoObj := types.ObjectValueMust(geo.AttributeTypes(), map[string]attr.Value{
		"mode":              geo.Mode,
		"countries":         geo.Countries,
		"traffic_direction": geo.TrafficDirection,
	})

	m := &usgModel{GeoIPFiltering: geoObj}
	body, diags := m.AsUnifiModel(ctx)
	require.False(t, diags.HasError(), "AsUnifiModel diags: %v", diags)

	usg, ok := body.(*unifi.SettingUsg)
	require.True(t, ok)
	assert.False(t, usg.GeoIPFilteringEnabled, "geo must NOT be enabled/written from unifi_setting_usg (deprecated no-op)")
	assert.Empty(t, usg.GeoIPFilteringCountries, "geo countries must not be written from unifi_setting_usg")

	// And Merge reports the enabled flag as a stable false.
	require.False(t, m.Merge(ctx, &unifi.SettingUsg{}).HasError())
	assert.Equal(t, types.BoolValue(false), m.GeoIPFilteringEnabled)
}
