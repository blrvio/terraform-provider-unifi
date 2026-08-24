package device

import (
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPortOverrideAggregateTranslation(t *testing.T) {
	t.Run("CountBecomesContiguousMemberRange", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{
			"number":              5,
			"name":                "lag uplink",
			"port_profile_id":     "",
			"op_mode":             "aggregate",
			"poe_mode":            "",
			"aggregate_num_ports": 3,
		})
		require.NoError(t, err)
		assert.Equal(t, []int{5, 6, 7}, po.AggregateMembers)
	})

	t.Run("UnsetLeavesMembersNil", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{
			"number":              1,
			"name":                "plain port",
			"port_profile_id":     "abc123",
			"op_mode":             "switch",
			"poe_mode":            "auto",
			"aggregate_num_ports": 0,
		})
		require.NoError(t, err)
		assert.Nil(t, po.AggregateMembers)
	})
}

func TestFromPortOverrideAggregateTranslation(t *testing.T) {
	t.Run("MemberListLengthBecomesCount", func(t *testing.T) {
		m := fromPortOverride(unifi.DevicePortOverrides{
			PortIDX:          5,
			OpMode:           "aggregate",
			AggregateMembers: []int{5, 6, 7},
		})
		assert.Equal(t, 3, m["aggregate_num_ports"])
	})

	t.Run("NilMembersReadBackAsZero", func(t *testing.T) {
		m := fromPortOverride(unifi.DevicePortOverrides{
			PortIDX: 1,
			OpMode:  "switch",
		})
		assert.Equal(t, 0, m["aggregate_num_ports"])
	})
}

func TestPortOverrideAggregateRoundTrip(t *testing.T) {
	in := map[string]interface{}{
		"number":              2,
		"name":                "lag",
		"port_profile_id":     "",
		"op_mode":             "aggregate",
		"poe_mode":            "",
		"aggregate_num_ports": 4,
	}
	po, err := toPortOverride(in)
	require.NoError(t, err)
	out := fromPortOverride(po)
	assert.Equal(t, in["aggregate_num_ports"], out["aggregate_num_ports"])
}

func stringSet(items ...string) *schema.Set {
	raw := make([]interface{}, len(items))
	for i, s := range items {
		raw[i] = s
	}
	return schema.NewSet(schema.HashString, raw)
}

// portOverrideData builds a complete port_override map (all schema keys present,
// as SDKv2 supplies them) so toPortOverride's type assertions don't panic.
func portOverrideData(overrides map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"number":                1,
		"name":                  "",
		"port_profile_id":       "",
		"op_mode":               "switch",
		"poe_mode":              "",
		"aggregate_num_ports":   0,
		"native_networkconf_id": "",
		"tagged_vlan_mgmt":      "",
		"forward":               "",
		"excluded_network_ids":  stringSet(),
		"voice_networkconf_id":  "",
		"setting_preference":    "",
	}
	for k, v := range overrides {
		data[k] = v
	}
	return data
}

func TestToPortOverride_VLANFields(t *testing.T) {
	po, err := toPortOverride(portOverrideData(map[string]interface{}{
		"number":                10,
		"native_networkconf_id": "net-native",
		"tagged_vlan_mgmt":      "custom",
		"forward":               "customize",
		"excluded_network_ids":  stringSet("net-a", "net-b"),
		"voice_networkconf_id":  "net-voice",
		"setting_preference":    "manual",
	}))
	require.NoError(t, err)
	assert.Equal(t, "net-native", po.NATiveNetworkID)
	assert.Equal(t, "custom", po.TaggedVLANMgmt)
	assert.Equal(t, "customize", po.Forward)
	assert.Equal(t, "net-voice", po.VoiceNetworkID)
	assert.Equal(t, "manual", po.SettingPreference)
	// TypeSet -> unordered slice: assert membership, not order.
	assert.ElementsMatch(t, []string{"net-a", "net-b"}, po.ExcludedNetworkIDs)
}

// forward has NO default: a block that sets only poe_mode must leave Forward
// empty (so omitempty drops it from the PUT and never black-holes a trunk port).
func TestToPortOverride_ForwardNoDefault(t *testing.T) {
	po, err := toPortOverride(portOverrideData(map[string]interface{}{
		"number":   3,
		"poe_mode": "auto",
	}))
	require.NoError(t, err)
	assert.Equal(t, "", po.Forward)
	assert.Equal(t, "", po.NATiveNetworkID)
	assert.Equal(t, "", po.TaggedVLANMgmt)
	assert.Equal(t, "", po.SettingPreference)
	assert.Empty(t, po.ExcludedNetworkIDs)
}

func TestFromPortOverride_VLANFields(t *testing.T) {
	m := fromPortOverride(unifi.DevicePortOverrides{
		PortIDX:            10,
		NATiveNetworkID:    "net-native",
		TaggedVLANMgmt:     "custom",
		Forward:            "customize",
		ExcludedNetworkIDs: []string{"net-a", "net-b"},
		VoiceNetworkID:     "net-voice",
		SettingPreference:  "manual",
	})
	assert.Equal(t, "net-native", m["native_networkconf_id"])
	assert.Equal(t, "custom", m["tagged_vlan_mgmt"])
	assert.Equal(t, "customize", m["forward"])
	assert.Equal(t, "net-voice", m["voice_networkconf_id"])
	assert.Equal(t, "manual", m["setting_preference"])
	excluded, ok := m["excluded_network_ids"].(*schema.Set)
	require.True(t, ok)
	got := make([]string, 0, excluded.Len())
	for _, v := range excluded.List() {
		s, _ := v.(string)
		got = append(got, s)
	}
	assert.ElementsMatch(t, []string{"net-a", "net-b"}, got)
}

func TestPortOverride_VLANRoundTrip(t *testing.T) {
	in := portOverrideData(map[string]interface{}{
		"number":                7,
		"native_networkconf_id": "net-native",
		"tagged_vlan_mgmt":      "custom",
		"forward":               "customize",
		"excluded_network_ids":  stringSet("net-a", "net-b"),
		"voice_networkconf_id":  "net-voice",
		"setting_preference":    "manual",
	})
	po, err := toPortOverride(in)
	require.NoError(t, err)
	out := fromPortOverride(po)
	assert.Equal(t, in["native_networkconf_id"], out["native_networkconf_id"])
	assert.Equal(t, in["tagged_vlan_mgmt"], out["tagged_vlan_mgmt"])
	assert.Equal(t, in["forward"], out["forward"])
	assert.Equal(t, in["voice_networkconf_id"], out["voice_networkconf_id"])
	assert.Equal(t, in["setting_preference"], out["setting_preference"])
	inExcluded, _ := in["excluded_network_ids"].(*schema.Set)
	outExcluded, _ := out["excluded_network_ids"].(*schema.Set)
	assert.True(t, inExcluded.Equal(outExcluded))
}

// The StringInSlice validators are the only client-side guard (SDK validation is
// disabled in base/client.go), so verify they accept valid and reject invalid
// values for the constrained VLAN attributes.
func TestPortOverride_VLANValidators(t *testing.T) {
	elem, _ := ResourceDevice().Schema["port_override"].Elem.(*schema.Resource)

	cases := []struct {
		attr    string
		valid   []string
		invalid []string
	}{
		{"tagged_vlan_mgmt", []string{"auto", "block_all", "custom"}, []string{"bogus", "all"}},
		{"forward", []string{"all", "native", "customize", "disabled"}, []string{"bogus", "trunk"}},
		{"setting_preference", []string{"auto", "manual"}, []string{"bogus", "automatic"}},
	}
	for _, tc := range cases {
		t.Run(tc.attr, func(t *testing.T) {
			vf := elem.Schema[tc.attr].ValidateFunc
			require.NotNil(t, vf, "%s must have a ValidateFunc", tc.attr)
			for _, v := range tc.valid {
				_, errs := vf(v, tc.attr)
				assert.Empty(t, errs, "%q should be valid for %s", v, tc.attr)
			}
			for _, v := range tc.invalid {
				_, errs := vf(v, tc.attr)
				assert.NotEmpty(t, errs, "%q should be rejected for %s", v, tc.attr)
			}
		})
	}
}

// portOverrideSetHash must key the set by port number ONLY, so a controller
// echoing/auto-populating a per-port VLAN field (e.g. setting_preference or a
// native VLAN) on an entry the user didn't declare it on cannot change the
// element's set identity and churn the set (the backward-compat hazard). Two
// blocks for the same port number must hash identically regardless of the VLAN
// fields; different numbers must hash differently.
func TestPortOverrideSetHash_StableByNumber(t *testing.T) {
	bare := portOverrideData(map[string]interface{}{"number": 5})
	echoed := portOverrideData(map[string]interface{}{
		"number":                5,
		"setting_preference":    "auto",
		"native_networkconf_id": "net-auto",
		"forward":               "native",
	})
	assert.Equal(t, portOverrideSetHash(bare), portOverrideSetHash(echoed),
		"same port number must hash identically regardless of echoed VLAN fields")

	other := portOverrideData(map[string]interface{}{"number": 6})
	assert.NotEqual(t, portOverrideSetHash(bare), portOverrideSetHash(other),
		"different port numbers must hash differently")

	// Sanity: the hash must be a pure function of `number`.
	assert.Equal(t, portOverrideSetHash(bare), portOverrideSetHash(bare))
}

// The new VLAN attributes must be Optional+Computed so an undeclared field reads
// back the controller's value without a perpetual diff. This pairs with the
// number-keyed set hash above; together they neutralize the upgrade churn.
func TestPortOverrideVLANFields_AreOptionalComputed(t *testing.T) {
	elem, _ := ResourceDevice().Schema["port_override"].Elem.(*schema.Resource)
	for _, attr := range []string{
		"native_networkconf_id", "tagged_vlan_mgmt", "forward",
		"excluded_network_ids", "voice_networkconf_id", "setting_preference",
	} {
		s := elem.Schema[attr]
		require.NotNil(t, s, "attribute %s must exist", attr)
		assert.True(t, s.Optional, "%s must be Optional", attr)
		assert.True(t, s.Computed, "%s must be Computed (to absorb controller echoes without churn)", attr)
		assert.Nil(t, s.Default, "%s must not have a Default", attr)
	}
}

func radioSet(items ...map[string]interface{}) *schema.Set {
	raw := make([]interface{}, len(items))
	for i, m := range items {
		raw[i] = m
	}
	return schema.NewSet(radioSetHash, raw)
}

func radioByBand(rs []unifi.DeviceRadioTable, band string) (unifi.DeviceRadioTable, bool) {
	for _, r := range rs {
		if r.Radio == band {
			return r, true
		}
	}
	return unifi.DeviceRadioTable{}, false
}

// The core safety property: declaring one band must not wipe the others.
func TestMergeRadios_PreservesUndeclaredBands(t *testing.T) {
	current := []unifi.DeviceRadioTable{
		{Radio: "ng", Channel: "1", Ht: 20, TxPowerMode: "auto"},
		{Radio: "na", Channel: "36", Ht: 80, TxPowerMode: "high"},
		{Radio: "6e", Channel: "37", Ht: 160, TxPowerMode: "auto"},
	}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "tx_power_mode": "disabled",
	}))
	if len(got) != 3 {
		t.Fatalf("expected all 3 bands preserved, got %d: %+v", len(got), got)
	}
	ng, _ := radioByBand(got, "ng")
	if ng.TxPowerMode != "disabled" {
		t.Errorf("ng tx_power_mode = %q, want disabled", ng.TxPowerMode)
	}
	if ng.Channel != "1" || ng.Ht != 20 {
		t.Errorf("ng channel/ht clobbered: got channel=%q ht=%d, want 1/20", ng.Channel, ng.Ht)
	}
	if na, _ := radioByBand(got, "na"); na.Channel != "36" || na.Ht != 80 || na.TxPowerMode != "high" {
		t.Errorf("na band modified: %+v", na)
	}
	if sixE, _ := radioByBand(got, "6e"); sixE.Channel != "37" || sixE.Ht != 160 {
		t.Errorf("6e band modified: %+v", sixE)
	}
}

// TestMergeRadios_Preserves5GCapabilityFields locks the fix for
// `400 api.err.DeviceNotSupport5gException` on a UAP6MP. Changing only the 2.4GHz
// (`ng`) radio must re-emit the 5GHz (`na`) radio with its capability fields
// (nss, radio_caps, radio_caps2, has_ht160, max/min_txpower) intact — the
// controller rejects a radio_table PUT that drops them. mergeRadios starts from
// the device's current radio_table, so once the SDK models these fields they
// survive the merge unchanged.
func TestMergeRadios_Preserves5GCapabilityFields(t *testing.T) {
	current := []unifi.DeviceRadioTable{
		{Radio: "ng", Channel: "auto", Ht: 20, TxPowerMode: "auto"},
		{
			Radio: "na", Channel: "40", Ht: 80, TxPowerMode: "auto",
			Nss: 4, RadioCaps: 251805700, RadioCaps2: 31,
			HasHt160: true, HasDfs: true, HasFccdfs: true,
			Is11Ac: true, Is11Ax: true, MaxTxpower: 26, MinTxpower: 6,
		},
	}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "channel": "1", "min_rssi": -75, "min_rssi_enabled": true,
	}))

	na, ok := radioByBand(got, "na")
	require.True(t, ok, "na band must be preserved")
	assert.Equal(t, 4, na.Nss, "nss dropped")
	assert.Equal(t, 251805700, na.RadioCaps, "radio_caps dropped")
	assert.Equal(t, 31, na.RadioCaps2, "radio_caps2 dropped")
	assert.True(t, na.HasHt160, "has_ht160 dropped")
	assert.True(t, na.Is11Ax, "is_11ax dropped")
	assert.Equal(t, 26, na.MaxTxpower, "max_txpower dropped")
	assert.Equal(t, 6, na.MinTxpower, "min_txpower dropped")

	ng, _ := radioByBand(got, "ng")
	assert.Equal(t, "1", ng.Channel, "ng channel must be applied")
}

// TestMergeRadiosRaw_PreservesAllFields is the byte-level regression for
// api.err.DeviceNotSupport5gException on a UAP6MP. It feeds a REAL 10.5.67
// radio_table (including a field the typed struct does not model and zero-valued
// fields it would drop), declares a change to the 2.4GHz radio only, and asserts
// that (a) the 5GHz band is preserved field-for-field and (b) the declared
// scalars are applied. This is the coverage the earlier typed-only unit test
// lacked, which is why it missed the real PUT failure.
func TestMergeRadiosRaw_PreservesAllFields(t *testing.T) {
	// Real na (5GHz) radio + ng (2.4GHz), plus an unmodeled field to prove
	// byte-preservation of things the provider does not know about.
	current := []json.RawMessage{
		json.RawMessage(`{"radio":"ng","channel":"auto","ht":20,"tx_power_mode":"auto","min_rssi_enabled":false,"current_antenna_gain":0,"nss":2,"radio_caps":123,"vendor_future_field":"keep-me"}`),
		json.RawMessage(`{"radio":"na","channel":40,"ht":"80","nss":4,"radio_caps":251805700,"radio_caps2":31,"has_ht160":true,"max_txpower":26,"min_txpower":6,"min_rssi_enabled":false,"current_antenna_gain":0,"vendor_future_field":"keep-me-too"}`),
	}

	merged, err := mergeRadiosRaw(current, radioSet(map[string]interface{}{
		"name": "ng", "channel": "1", "min_rssi": -75, "min_rssi_enabled": true,
	}))
	require.NoError(t, err)
	require.Len(t, merged, 2)

	byBand := map[string]map[string]json.RawMessage{}
	for _, raw := range merged {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &m))
		var b struct {
			Radio string `json:"radio"`
		}
		require.NoError(t, json.Unmarshal(raw, &b))
		byBand[b.Radio] = m
	}

	// (a) na band: every original field preserved byte-for-byte, nothing dropped.
	na := byBand["na"]
	for k, want := range map[string]string{
		"nss": "4", "radio_caps": "251805700", "radio_caps2": "31", "has_ht160": "true",
		"max_txpower": "26", "min_txpower": "6", "min_rssi_enabled": "false",
		"current_antenna_gain": "0", "channel": "40", "ht": `"80"`,
		"vendor_future_field": `"keep-me-too"`,
	} {
		got, ok := na[k]
		require.Truef(t, ok, "na.%s must be preserved", k)
		assert.Equalf(t, want, string(got), "na.%s must be byte-preserved", k)
	}

	// (b) ng band: declared scalars applied; unmodeled/other fields preserved.
	ng := byBand["ng"]
	assert.Equal(t, `"1"`, string(ng["channel"]), "ng.channel must be applied")
	assert.Equal(t, "-75", string(ng["min_rssi"]), "ng.min_rssi must be applied")
	assert.Equal(t, "true", string(ng["min_rssi_enabled"]), "ng.min_rssi_enabled must be applied")
	assert.Equal(t, `"keep-me"`, string(ng["vendor_future_field"]), "ng unmodeled field must be preserved")
	assert.Equal(t, "123", string(ng["radio_caps"]), "ng.radio_caps must be preserved")
}

// Only non-zero declared fields overlay; unset fields keep the controller value.
func TestMergeRadios_OverlaysOnlyNonZero(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "ng", Channel: "6", Ht: 40, TxPowerMode: "medium"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "channel": "11",
	}))
	ng, _ := radioByBand(got, "ng")
	if ng.Channel != "11" {
		t.Errorf("channel = %q, want 11", ng.Channel)
	}
	if ng.Ht != 40 || ng.TxPowerMode != "medium" {
		t.Errorf("unset fields clobbered: ht=%d tx_power_mode=%q, want 40/medium", ng.Ht, ng.TxPowerMode)
	}
}

// A declared band missing from the device is appended.
func TestMergeRadios_AppendsMissingBand(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "na", Channel: "36"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "tx_power_mode": "disabled",
	}))
	if len(got) != 2 {
		t.Fatalf("expected 2 bands, got %d", len(got))
	}
	if ng, ok := radioByBand(got, "ng"); !ok || ng.TxPowerMode != "disabled" {
		t.Errorf("ng not appended correctly: %+v ok=%v", ng, ok)
	}
}

// min_rssi pairs with min_rssi_enabled only when a non-zero threshold is set.
func TestMergeRadios_MinRssiPairing(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "na"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "na", "min_rssi": -75, "min_rssi_enabled": true,
	}))
	na, _ := radioByBand(got, "na")
	if na.MinRssi != -75 || !na.MinRssiEnabled {
		t.Errorf("min_rssi pairing failed: %+v", na)
	}
}

// Etherlighting overlay: declared fields apply; unset fields keep current values.
func TestMergeEtherLighting_OverlaysOnlyNonZero(t *testing.T) {
	current := unifi.DeviceEtherLighting{Mode: "speed", Brightness: 80, Behavior: "steady", LedMode: "etherlighting"}
	got := mergeEtherLighting(current, map[string]interface{}{"mode": "network"})
	if got.Mode != "network" {
		t.Errorf("mode = %q, want network", got.Mode)
	}
	if got.Brightness != 80 || got.Behavior != "steady" || got.LedMode != "etherlighting" {
		t.Errorf("unset fields clobbered: %+v", got)
	}
}

func TestMergeEtherLighting_FromEmptyCurrent(t *testing.T) {
	got := mergeEtherLighting(unifi.DeviceEtherLighting{}, map[string]interface{}{"mode": "network", "brightness": 60})
	if got.Mode != "network" || got.Brightness != 60 || got.Behavior != "" {
		t.Errorf("unexpected merge from empty: %+v", got)
	}
}
