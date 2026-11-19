package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

// G7: importing unifi_setting_usg starts from a zero-value model; Merge must leave
// geo_ip_filtering as a typed null so the state can be set (it used to fail with
// "Value Conversion Error").
func TestUsgImportStateSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var sresp resource.SchemaResponse
	NewUsgResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	require.False(t, sresp.Diagnostics.HasError())

	m := &usgModel{}
	require.False(t, m.Merge(ctx, &unifi.SettingUsg{ID: "u1", UpnpEnabled: false}).HasError())
	m.SetSite("default")

	state := tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	diags := state.Set(ctx, m)
	require.False(t, diags.HasError(), "state.Set: %v", diags)
	assert.True(t, m.GeoIPFiltering.IsNull())
}
