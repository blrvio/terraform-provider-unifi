package wifibroadcast

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// marshalBody runs a model through AsUnifiModel and marshals the resulting
// Official-API body to JSON, exercising the plain-struct/MarshalJSON overlay.
func marshalBody(t *testing.T, m *wifiBroadcastModel) map[string]any {
	t.Helper()
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestWifiBroadcastAsUnifiModel_Standard(t *testing.T) {
	t.Parallel()
	m := &wifiBroadcastModel{
		Name:                  types.StringValue("Corp-WiFi"),
		Type:                  types.StringValue(wifiBroadcastTypeStandard),
		SecurityConfiguration: types.StringValue(`{"type":"WPA2_PERSONAL","passphrase":"supersecret"}`),
		// Enabled left null -> should default to true.
	}
	out := marshalBody(t, m)
	assert.Equal(t, "Corp-WiFi", out["name"])
	assert.Equal(t, "STANDARD", out["type"])
	assert.Equal(t, true, out["enabled"], "enabled must survive the top-level MarshalJSON overlay")
	// Bool defaults must be present on the wire (all false).
	assert.Equal(t, false, out["hideName"])
	assert.Equal(t, false, out["clientIsolationEnabled"])

	// security_configuration must be decoded into the union and re-emitted as an object.
	sec, ok := out["securityConfiguration"].(map[string]any)
	require.True(t, ok, "securityConfiguration should be a JSON object, got %T", out["securityConfiguration"])
	assert.Equal(t, "WPA2_PERSONAL", sec["type"])
	assert.Equal(t, "supersecret", sec["passphrase"])
}

func TestWifiBroadcastAsUnifiModel_DisabledIoT(t *testing.T) {
	t.Parallel()
	m := &wifiBroadcastModel{
		Name:                  types.StringValue("IoT-WiFi"),
		Type:                  types.StringValue(wifiBroadcastTypeIoT),
		Enabled:               types.BoolValue(false),
		HideName:              types.BoolValue(true),
		SecurityConfiguration: types.StringValue(`{"type":"OPEN"}`),
		ClientFilteringPolicy: types.StringValue(`{"action":"BLOCK","macAddressFilter":["aa:bb:cc:dd:ee:ff"]}`),
	}
	out := marshalBody(t, m)
	assert.Equal(t, "IOT_OPTIMIZED", out["type"])
	assert.Equal(t, false, out["enabled"])
	assert.Equal(t, true, out["hideName"])

	// Optional nested JSON attribute must be decoded and present on the wire.
	pol, ok := out["clientFilteringPolicy"].(map[string]any)
	require.True(t, ok, "clientFilteringPolicy should be a JSON object, got %T", out["clientFilteringPolicy"])
	assert.Equal(t, "BLOCK", pol["action"])

	// Unset optional nested objects must be omitted (pointer left nil).
	_, present := out["network"]
	assert.False(t, present, "network should be omitted when unset")
}

func TestWifiBroadcastAsUnifiModel_MissingSecurity(t *testing.T) {
	t.Parallel()
	m := &wifiBroadcastModel{
		Name: types.StringValue("No-Sec"),
		Type: types.StringValue(wifiBroadcastTypeStandard),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}

func TestWifiBroadcastAsUnifiModel_InvalidNestedJSON(t *testing.T) {
	t.Parallel()
	m := &wifiBroadcastModel{
		Name:                  types.StringValue("Bad-JSON"),
		Type:                  types.StringValue(wifiBroadcastTypeStandard),
		SecurityConfiguration: types.StringValue(`{"type":"OPEN"}`),
		Network:               types.StringValue(`{not valid json`),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}

func TestWifiBroadcastMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	// Simulate a controller read response.
	raw := `{"id":"11111111-2222-3333-4444-555555555555","type":"STANDARD","enabled":true,` +
		`"name":"Corp-WiFi","hideName":false,"clientIsolationEnabled":true,"uapsdEnabled":false,` +
		`"multicastToUnicastConversionEnabled":false,"channel2gLockedTo6":false,"dtimPeriod2gLockedTo3":false,` +
		`"securityConfiguration":{"type":"WPA2_PERSONAL","passphrase":"supersecret"},` +
		`"clientFilteringPolicy":{"action":"BLOCK","macAddressFilter":["aa:bb:cc:dd:ee:ff"]}}`
	var details official.WifiBroadcastDetails
	require.NoError(t, json.Unmarshal([]byte(raw), &details))

	m := &wifiBroadcastModel{}
	diags := m.Merge(context.Background(), &details)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "11111111-2222-3333-4444-555555555555", m.GetID())
	assert.Equal(t, "Corp-WiFi", m.Name.ValueString())
	assert.Equal(t, "STANDARD", m.Type.ValueString())
	assert.True(t, m.Enabled.ValueBool())
	assert.True(t, m.ClientIsolationEnabled.ValueBool())
	assert.False(t, m.HideName.ValueBool())

	// security_configuration must round-trip to a non-empty JSON string.
	require.False(t, m.SecurityConfiguration.IsNull())
	sec := map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(m.SecurityConfiguration.ValueString()), &sec))
	assert.Equal(t, "WPA2_PERSONAL", sec["type"])

	// Optional nested object present -> string set.
	require.False(t, m.ClientFilteringPolicy.IsNull())
	// Optional nested object absent -> null.
	assert.True(t, m.Network.IsNull())
	assert.True(t, m.MdnsProxy.IsNull())
}

func TestWifiBroadcastMerge_EnabledSurvives(t *testing.T) {
	t.Parallel()
	// Confirm the top-level `enabled` overlay is read back correctly (false).
	raw := `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","type":"IOT_OPTIMIZED","enabled":false,` +
		`"name":"IoT-WiFi","securityConfiguration":{"type":"OPEN"}}`
	var details official.WifiBroadcastDetails
	require.NoError(t, json.Unmarshal([]byte(raw), &details))

	m := &wifiBroadcastModel{}
	diags := m.Merge(context.Background(), &details)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "IOT_OPTIMIZED", m.Type.ValueString())
	assert.False(t, m.Enabled.ValueBool())
	assert.Equal(t, "IoT-WiFi", m.Name.ValueString())
}
