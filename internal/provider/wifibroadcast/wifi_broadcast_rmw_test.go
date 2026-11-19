package wifibroadcast

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// currentStandardJSON mirrors a GET of a STANDARD broadcast on UniFi Network
// 11.0.99, including variant fields the resource does not model.
const currentStandardJSON = `{
  "id": "cdc8081d-a343-4787-bfc9-296564a7b220",
  "metadata": {"origin": "USER_DEFINED"},
  "type": "STANDARD",
  "name": "old-name",
  "enabled": true,
  "hideName": false,
  "clientIsolationEnabled": false,
  "uapsdEnabled": false,
  "multicastToUnicastConversionEnabled": true,
  "channel2gLockedTo6": false,
  "dtimPeriod2gLockedTo3": false,
  "securityConfiguration": {"type": "WPA2_PERSONAL", "passphrase": "old-secret"},
  "network": {"type": "NATIVE"},
  "clientFilteringPolicy": {"action": "BLOCK", "macAddressFilter": ["aa:bb:cc:dd:ee:ff"]},
  "mdnsProxyConfiguration": {"mode": "AUTO"},
  "arpProxyEnabled": false,
  "bssTransitionEnabled": true,
  "advertiseDeviceName": false,
  "bandSteeringEnabled": false,
  "mloEnabled": false,
  "broadcastingFrequenciesGHz": [2.4, 5],
  "hotspotConfiguration": {"type": "NONE"}
}`

func overlayFromModel(t *testing.T, m *wifiBroadcastModel) map[string]any {
	t.Helper()
	current := &official.WifiBroadcastDetails{}
	require.NoError(t, json.Unmarshal([]byte(currentStandardJSON), current))

	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	planBody, diags := asWifiBroadcastBody(body)
	require.False(t, diags.HasError())

	merged, err := overlayWifiBroadcast(current, planBody)
	require.NoError(t, err)
	raw, err := json.Marshal(merged)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestOverlayWifiBroadcast(t *testing.T) {
	t.Parallel()
	out := overlayFromModel(t, &wifiBroadcastModel{
		Name:                  types.StringValue("mashiro-fifi"),
		Type:                  types.StringValue(wifiBroadcastTypeStandard),
		Enabled:               types.BoolValue(true),
		UapsdEnabled:          types.BoolValue(true),
		SecurityConfiguration: ut.JSONStringValue(`{"type":"WPA2_PERSONAL","passphrase":"new-secret"}`),
		Network:               ut.JSONStringValue(`{"type":"NATIVE"}`),
		BlackoutSchedule:      ut.JSONStringValue(`{"days":[{"day":"MON","type":"ALL_DAY"}]}`),
		// client_filtering_policy and mdns_proxy null: must be removed, not kept.
		ClientFilteringPolicy: ut.JSONNull(),
		MdnsProxy:             ut.JSONNull(),
	})

	t.Run("read-only keys dropped", func(t *testing.T) {
		t.Parallel()
		assert.NotContains(t, out, "id")
		assert.NotContains(t, out, "metadata")
	})
	t.Run("unmodeled variant fields kept", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, false, out["arpProxyEnabled"])
		assert.Equal(t, true, out["bssTransitionEnabled"])
		assert.Equal(t, false, out["advertiseDeviceName"])
		assert.Equal(t, []any{2.4, float64(5)}, out["broadcastingFrequenciesGHz"])
		assert.Equal(t, map[string]any{"type": "NONE"}, out["hotspotConfiguration"])
	})
	t.Run("plan fields win", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "mashiro-fifi", out["name"])
		assert.Equal(t, true, out["uapsdEnabled"])
		assert.Equal(t, false, out["multicastToUnicastConversionEnabled"], "plan default false overrides current true")
		sec, ok := out["securityConfiguration"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "WPA2_PERSONAL", sec["type"])
		assert.Equal(t, "new-secret", sec["passphrase"])
		assert.Equal(t, map[string]any{"days": []any{map[string]any{"day": "MON", "type": "ALL_DAY"}}}, out["blackoutScheduleConfiguration"])
	})
	t.Run("null managed objects removed", func(t *testing.T) {
		t.Parallel()
		assert.NotContains(t, out, "clientFilteringPolicy")
		assert.NotContains(t, out, "mdnsProxyConfiguration")
	})
}

func TestOverlayWifiBroadcast_SetsClientFilter(t *testing.T) {
	t.Parallel()
	out := overlayFromModel(t, &wifiBroadcastModel{
		Name:                  types.StringValue("mashiro-fifi"),
		Type:                  types.StringValue(wifiBroadcastTypeStandard),
		SecurityConfiguration: ut.JSONStringValue(`{"type":"WPA2_PERSONAL","passphrase":"x"}`),
		ClientFilteringPolicy: ut.JSONStringValue(`{"action":"ALLOW","macAddressFilter":["00:11:22:33:44:55","66:77:88:99:aa:bb"]}`),
	})
	assert.Equal(t, map[string]any{
		"action":           "ALLOW",
		"macAddressFilter": []any{"00:11:22:33:44:55", "66:77:88:99:aa:bb"},
	}, out["clientFilteringPolicy"])
	assert.NotContains(t, out, "network", "network null in plan is removed")
}
