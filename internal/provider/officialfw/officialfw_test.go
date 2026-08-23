package officialfw

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listOfStrings builds a Framework list value from the given strings for tests.
func listOfStrings(t *testing.T, in ...string) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), types.StringType, in)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

// ---- Zone ----

func TestFirewallZoneAsUnifiModel_NetworkIDs(t *testing.T) {
	t.Parallel()
	m := &officialFirewallZoneModel{
		Name: types.StringValue("LAN"),
		NetworkIDs: listOfStrings(t,
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
		),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	zoneBody, ok := body.(*official.FirewallZoneCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallZoneCreateOrUpdate, got %T", body)
	assert.Equal(t, "LAN", zoneBody.Name)
	require.Len(t, zoneBody.NetworkIds, 2)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", zoneBody.NetworkIds[0].String())
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", zoneBody.NetworkIds[1].String())

	// Verify the wire representation carries the camelCase key.
	raw, err := json.Marshal(zoneBody)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, "LAN", out["name"])
	assert.Len(t, out["networkIds"], 2)
}

func TestFirewallZoneAsUnifiModel_InvalidUUID(t *testing.T) {
	t.Parallel()
	m := &officialFirewallZoneModel{
		Name:       types.StringValue("LAN"),
		NetworkIDs: listOfStrings(t, "not-a-uuid"),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError(), "expected error for invalid UUID")
}

func TestFirewallZoneMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"33333333-3333-3333-3333-333333333333","name":"IoT",` +
		`"networkIds":["44444444-4444-4444-4444-444444444444","55555555-5555-5555-5555-555555555555"]}`
	var zone official.FirewallZone
	require.NoError(t, json.Unmarshal([]byte(raw), &zone))

	m := &officialFirewallZoneModel{}
	diags := m.Merge(context.Background(), &zone)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "33333333-3333-3333-3333-333333333333", m.GetID())
	assert.Equal(t, "IoT", m.Name.ValueString())

	var got []string
	m.NetworkIDs.ElementsAs(context.Background(), &got, false)
	assert.Equal(t, []string{
		"44444444-4444-4444-4444-444444444444",
		"55555555-5555-5555-5555-555555555555",
	}, got)
}

// ---- Policy ----

func TestFirewallPolicyAsUnifiModel_Body(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:            types.StringValue("allow-lan-to-wan"),
		Action:          types.StringValue(firewallPolicyActionAllow),
		Source:          types.StringValue(`{"zoneId":"11111111-1111-1111-1111-111111111111"}`),
		Destination:     types.StringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope: types.StringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
		// Enabled null -> defaults true; LoggingEnabled null -> defaults false.
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	policyBody, ok := body.(*official.FirewallPolicyCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallPolicyCreateOrUpdate, got %T", body)
	assert.Equal(t, "allow-lan-to-wan", policyBody.Name)
	assert.Equal(t, firewallPolicyActionAllow, policyBody.Action.Type)
	assert.True(t, policyBody.Enabled)
	assert.False(t, policyBody.LoggingEnabled)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", policyBody.Source.ZoneId.String())
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", policyBody.Destination.ZoneId.String())

	// Verify the wire keys on the marshaled body.
	raw, err := json.Marshal(policyBody)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, "allow-lan-to-wan", out["name"])
	assert.Equal(t, true, out["enabled"])
	action, ok := out["action"].(map[string]any)
	require.True(t, ok, "action should be an object")
	assert.Equal(t, "ALLOW", action["type"])
	source, ok := out["source"].(map[string]any)
	require.True(t, ok, "source should be an object")
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", source["zoneId"])
	dest, ok := out["destination"].(map[string]any)
	require.True(t, ok, "destination should be an object")
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", dest["zoneId"])
}

func TestFirewallPolicyAsUnifiModel_WithFilters(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:                  types.StringValue("block-invalid"),
		Action:                types.StringValue(firewallPolicyActionBlock),
		Enabled:               types.BoolValue(false),
		LoggingEnabled:        types.BoolValue(true),
		Description:           types.StringValue("drop invalid"),
		IpsecFilter:           types.StringValue(firewallPolicyIpsecMatchEncrypted),
		ConnectionStateFilter: listOfStrings(t, firewallPolicyConnStateNew, firewallPolicyConnStateInvalid),
		Source:                types.StringValue(`{"zoneId":"11111111-1111-1111-1111-111111111111"}`),
		Destination:           types.StringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope:       types.StringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	policyBody, ok := body.(*official.FirewallPolicyCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallPolicyCreateOrUpdate, got %T", body)
	assert.False(t, policyBody.Enabled)
	assert.True(t, policyBody.LoggingEnabled)
	require.NotNil(t, policyBody.Description)
	assert.Equal(t, "drop invalid", *policyBody.Description)
	require.NotNil(t, policyBody.IpsecFilter)
	assert.Equal(t, official.FirewallPolicyIpsecFilter(firewallPolicyIpsecMatchEncrypted), *policyBody.IpsecFilter)
	require.NotNil(t, policyBody.ConnectionStateFilter)
	assert.Equal(t, []official.FirewallPolicyConnectionStateFilter{
		official.FirewallPolicyConnectionStateFilter(firewallPolicyConnStateNew),
		official.FirewallPolicyConnectionStateFilter(firewallPolicyConnStateInvalid),
	}, *policyBody.ConnectionStateFilter)
}

func TestFirewallPolicyAsUnifiModel_InvalidSourceJSON(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:            types.StringValue("bad"),
		Action:          types.StringValue(firewallPolicyActionAllow),
		Source:          types.StringValue(`{not json`),
		Destination:     types.StringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope: types.StringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError(), "expected error for invalid source JSON")
}

func TestFirewallPolicyMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"66666666-6666-6666-6666-666666666666","name":"allow-lan-to-wan",` +
		`"action":{"type":"ALLOW"},"enabled":true,"loggingEnabled":false,"index":3,` +
		`"source":{"zoneId":"11111111-1111-1111-1111-111111111111"},` +
		`"destination":{"zoneId":"22222222-2222-2222-2222-222222222222"},` +
		`"ipProtocolScope":{"ipVersion":"BOTH","type":"ALL"},` +
		`"metadata":{}}`
	var policy official.FirewallPolicy
	require.NoError(t, json.Unmarshal([]byte(raw), &policy))

	m := &officialFirewallPolicyModel{}
	diags := m.Merge(context.Background(), &policy)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "66666666-6666-6666-6666-666666666666", m.GetID())
	assert.Equal(t, "allow-lan-to-wan", m.Name.ValueString())
	assert.Equal(t, "ALLOW", m.Action.ValueString())
	assert.True(t, m.Enabled.ValueBool())
	assert.False(t, m.LoggingEnabled.ValueBool())
	assert.Equal(t, int32(3), m.Index.ValueInt32())
	assert.True(t, m.IpsecFilter.IsNull())
	assert.True(t, m.ConnectionStateFilter.IsNull())
	assert.True(t, m.Schedule.IsNull())

	// The complex objects come back as compact JSON strings carrying the zoneId.
	assert.JSONEq(t, `{"zoneId":"11111111-1111-1111-1111-111111111111"}`, m.Source.ValueString())
	assert.JSONEq(t, `{"zoneId":"22222222-2222-2222-2222-222222222222"}`, m.Destination.ValueString())
	assert.JSONEq(t, `{"ipVersion":"BOTH","type":"ALL"}`, m.IPProtocolScope.ValueString())
}

// ---- Policy order ----

func TestFirewallPolicyOrderAsUnifiModel_Conversion(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyOrderModel{
		SourceZoneID:      types.StringValue("11111111-1111-1111-1111-111111111111"),
		DestinationZoneID: types.StringValue("22222222-2222-2222-2222-222222222222"),
		PolicyIDs: listOfStrings(t,
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	ordering, ok := body.(*official.FirewallPolicyOrdering)
	require.True(t, ok, "expected *official.FirewallPolicyOrdering, got %T", body)
	require.Len(t, ordering.OrderedFirewallPolicyIds.AfterSystemDefined, 2)
	assert.Equal(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", ordering.OrderedFirewallPolicyIds.AfterSystemDefined[0].String())
	assert.Equal(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", ordering.OrderedFirewallPolicyIds.AfterSystemDefined[1].String())
}

func TestFirewallPolicyOrderMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"orderedFirewallPolicyIds":{"afterSystemDefined":[` +
		`"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"],` +
		`"beforeSystemDefined":[]}}`
	var ordering official.FirewallPolicyOrdering
	require.NoError(t, json.Unmarshal([]byte(raw), &ordering))

	m := &officialFirewallPolicyOrderModel{}
	diags := m.Merge(context.Background(), &ordering)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	var got []string
	m.PolicyIDs.ElementsAs(context.Background(), &got, false)
	assert.Equal(t, []string{
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	}, got)
}

func TestFirewallPolicyOrderMerge_InvalidType(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyOrderModel{}
	diags := m.Merge(context.Background(), "not an ordering")
	assert.True(t, diags.HasError(), "expected error for wrong type")
}
