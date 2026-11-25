package device

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

// portOverrideSetHash must hash every declared field (G22): an in-element edit
// has to change the hash, otherwise the SDKv2 set diff sees equal sets and plans
// "No changes". Zero values, an omitted field and op_mode "switch" vs "" must
// hash identically, and nested sets must not depend on element order.
func TestPortOverrideSetHash_DeclaredFields(t *testing.T) {
	bare := portOverrideData(map[string]interface{}{"number": 5})

	assert.NotEqual(t, portOverrideSetHash(bare),
		portOverrideSetHash(portOverrideData(map[string]interface{}{"number": 6})),
		"different port numbers must hash differently")
	for field, value := range map[string]interface{}{
		"name":                  "Uplink",
		"forward":               "native",
		"native_networkconf_id": "net-main",
		"setting_preference":    "manual",
		"poe_mode":              "off",
		"excluded_network_ids":  stringSet("net-a"),
	} {
		edited := portOverrideData(map[string]interface{}{"number": 5, field: value})
		assert.NotEqual(t, portOverrideSetHash(bare), portOverrideSetHash(edited),
			"editing %s must change the element hash", field)
	}

	assert.Equal(t, portOverrideSetHash(bare), portOverrideSetHash(map[string]interface{}{"number": 5}),
		"zero values and omitted fields must hash identically")
	assert.Equal(t, portOverrideSetHash(bare),
		portOverrideSetHash(portOverrideData(map[string]interface{}{"number": 5, "op_mode": ""})),
		"op_mode \"switch\" (config default) and \"\" (controller) must hash identically")
	assert.Equal(t,
		portOverrideSetHash(portOverrideData(map[string]interface{}{"number": 5, "excluded_network_ids": stringSet("a", "b")})),
		portOverrideSetHash(portOverrideData(map[string]interface{}{"number": 5, "excluded_network_ids": stringSet("b", "a")})),
		"nested set order must not change the hash")
}

func TestRadioSetHash_DeclaredFields(t *testing.T) {
	base := map[string]interface{}{"name": "na", "channel": "149", "ht": 80, "tx_power_mode": "", "tx_power": "", "min_rssi": 0, "min_rssi_enabled": false}
	assert.Equal(t, radioSetHash(base), radioSetHash(map[string]interface{}{"name": "na", "channel": "149", "ht": 80}),
		"zero values and omitted fields must hash identically")
	for field, value := range map[string]interface{}{"channel": "36", "ht": 40, "tx_power_mode": "medium", "min_rssi_enabled": true, "min_rssi": -75} {
		edited := map[string]interface{}{}
		for k, v := range base {
			edited[k] = v
		}
		edited[field] = value
		assert.NotEqual(t, radioSetHash(base), radioSetHash(edited), "editing %s must change the element hash", field)
	}
	assert.NotEqual(t, radioSetHash(map[string]interface{}{"name": "na"}), radioSetHash(map[string]interface{}{"name": "ng"}))
}

// The read path keeps only the fields the prior state/config declares on a
// managed element, so a controller echo cannot move the hash, while drift on a
// declared field still surfaces. Unmanaged ports (console-made overrides,
// import) are read in full.
func TestSetFromPortOverrides_DeclaredFieldsOnly(t *testing.T) {
	controller := []unifi.DevicePortOverrides{
		{PortIDX: 16, Name: "MacBook (hub Dell)", Forward: "native", NATiveNetworkID: "net-main", SettingPreference: "manual", PoeMode: "auto"},
		{PortIDX: 1, Name: "Console-made", Forward: "all"},
	}
	prior := map[string]map[string]interface{}{
		"16": portOverrideData(map[string]interface{}{"number": 16, "name": "old name", "forward": "native", "native_networkconf_id": "net-main"}),
	}
	got := setFromPortOverrides(controller, prior)
	require.Len(t, got, 2)

	managed, unmanaged := got[0], got[1]
	assert.Equal(t, "MacBook (hub Dell)", managed["name"], "drift on a declared field must be read back")
	assert.Equal(t, "native", managed["forward"])
	assert.Equal(t, "net-main", managed["native_networkconf_id"])
	assert.Equal(t, "", managed["setting_preference"], "undeclared echoed field must be dropped")
	assert.Equal(t, "", managed["poe_mode"], "undeclared echoed field must be dropped")
	assert.Equal(t, portOverrideSetHash(portOverrideData(map[string]interface{}{
		"number": 16, "name": "MacBook (hub Dell)", "forward": "native", "native_networkconf_id": "net-main",
	})), portOverrideSetHash(managed), "state element must hash like the matching config element")

	assert.Equal(t, "Console-made", unmanaged["name"], "unmanaged override must be read in full")
	assert.Equal(t, "all", unmanaged["forward"])
}

// TestDeviceSetInElementEditsPlan is the end-to-end regression for G22: an edit
// inside an existing radio / port_override element must produce a diff, and a
// config that matches state must not.
func TestDeviceSetInElementEditsPlan(t *testing.T) {
	r := ResourceDevice()
	radioState := map[string]interface{}{"name": "na", "channel": "149", "ht": 80}
	portState := portOverrideData(map[string]interface{}{"number": 16, "name": "Port 16", "forward": "native"})
	rp := fmt.Sprintf("radio.%d.", radioSetHash(radioState))
	pp := fmt.Sprintf("port_override.%d.", portOverrideSetHash(portState))
	state := &terraform.InstanceState{ID: "dev1", Attributes: map[string]string{
		"id": "dev1", "mac": "ac:8b:a9:67:26:45", "name": "ap", "allow_adoption": "true", "forget_on_destroy": "true",
		"radio.#": "1", rp + "name": "na", rp + "channel": "149", rp + "ht": "80", rp + "tx_power_mode": "",
		rp + "tx_power": "", rp + "min_rssi": "0", rp + "min_rssi_enabled": "false",
		"port_override.#": "1", pp + "number": "16", pp + "name": "Port 16", pp + "op_mode": "switch", pp + "forward": "native",
		pp + "excluded_network_ids.#": "0",
	}}
	config := func(radio, port map[string]interface{}) *terraform.ResourceConfig {
		return terraform.NewResourceConfigRaw(map[string]interface{}{
			"mac": "ac:8b:a9:67:26:45", "name": "ap", "allow_adoption": true, "forget_on_destroy": true,
			"radio": []interface{}{radio}, "port_override": []interface{}{port},
		})
	}
	changed := func(radio, port map[string]interface{}) bool {
		d, err := r.SimpleDiff(context.Background(), state, config(radio, port), nil)
		require.NoError(t, err)
		return d != nil && len(d.Attributes) > 0
	}
	sameRadio := map[string]interface{}{"name": "na", "channel": "149", "ht": 80}
	samePort := map[string]interface{}{"number": 16, "name": "Port 16", "forward": "native"}

	assert.False(t, changed(sameRadio, samePort), "matching config must plan no changes")
	assert.True(t, changed(map[string]interface{}{"name": "na", "channel": "149", "ht": 80, "tx_power_mode": "medium"}, samePort),
		"radio tx_power_mode edit must plan a change")
	assert.True(t, changed(map[string]interface{}{"name": "na", "channel": "36", "ht": 80}, samePort),
		"radio channel edit must plan a change")
	assert.True(t, changed(sameRadio, map[string]interface{}{"number": 16, "name": "Dock", "forward": "native"}),
		"port_override name edit must plan a change")
	assert.True(t, changed(sameRadio, map[string]interface{}{"number": 16, "name": "Port 16", "forward": "all"}),
		"port_override forward edit must plan a change")
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

// TestMergeRadiosRaw_MatchesUIShape locks the fix for
// `400 api.err.DeviceNotSupport5gException` on a UAP6MP. The controller accepts a
// radio_table PUT only in the exact shape the UniFi UI sends (captured from a
// working request): the read-only capability fields (nss, radio_caps, has_ht160,
// ...) must be DROPPED and channel/ht must be JSON NUMBERS. Earlier attempts that
// preserved every field, or sent channel/ht as strings, were rejected — this
// asserts the shape that actually round-trips, including an undeclared band.
func TestMergeRadiosRaw_MatchesUIShape(t *testing.T) {
	current := []json.RawMessage{
		json.RawMessage(`{"radio":"ng","name":"wifi0","channel":"auto","ht":20,"tx_power_mode":"auto","antenna_id":-1,"antenna_gain":4,"vwire_enabled":true,"min_rssi_enabled":false,"nss":2,"radio_caps":123,"has_ht160":true}`),
		json.RawMessage(`{"radio":"na","name":"wifi1","channel":40,"ht":"80","tx_power_mode":"auto","antenna_id":-1,"antenna_gain":6,"vwire_enabled":true,"min_rssi_enabled":false,"nss":4,"radio_caps":251805700,"has_ht160":true,"max_txpower":26}`),
	}
	merged, err := mergeRadiosRaw(current, radioSet(map[string]interface{}{
		"name": "ng", "channel": "1", "ht": 20, "tx_power_mode": "auto",
		"min_rssi": -75, "min_rssi_enabled": true,
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

	// (a) declared ng band: capability dropped, editable kept, scalars as numbers.
	ng := byBand["ng"]
	for _, k := range []string{"nss", "radio_caps", "has_ht160"} {
		_, ok := ng[k]
		assert.Falsef(t, ok, "ng.%s (read-only capability) must be dropped", k)
	}
	for _, k := range []string{"name", "antenna_id", "antenna_gain", "vwire_enabled"} {
		_, ok := ng[k]
		assert.Truef(t, ok, "ng.%s must be kept (the UI sends it)", k)
	}
	assert.Equal(t, "1", string(ng["channel"]), "ng.channel must be JSON number 1")
	assert.Equal(t, "20", string(ng["ht"]), "ng.ht must be JSON number 20")
	assert.Equal(t, "-75", string(ng["min_rssi"]))
	assert.Equal(t, "true", string(ng["min_rssi_enabled"]))

	// (b) undeclared na band: preserved editable fields, capability dropped.
	na := byBand["na"]
	assert.Equal(t, "40", string(na["channel"]), "na.channel preserved as number")
	assert.Equal(t, `"80"`, string(na["ht"]), "na.ht preserved")
	_, hasNss := na["nss"]
	assert.False(t, hasNss, "na capability fields must be dropped")
}

// min_rssi is sent only when enabled; disabling drops the value.
func TestMergeRadiosRaw_MinRssiPairing(t *testing.T) {
	current := []json.RawMessage{json.RawMessage(`{"radio":"na","channel":36,"min_rssi":-80,"min_rssi_enabled":true}`)}
	merged, err := mergeRadiosRaw(current, radioSet(map[string]interface{}{
		"name": "na", "min_rssi_enabled": false,
	}))
	require.NoError(t, err)
	byBand := map[string]map[string]json.RawMessage{}
	for _, raw := range merged {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &m))
		byBand["na"] = m
	}
	assert.Equal(t, "false", string(byBand["na"]["min_rssi_enabled"]))
	_, ok := byBand["na"]["min_rssi"]
	assert.False(t, ok, "min_rssi must be dropped when disabled")
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

// G19: mirror_port_idx round-trips and is validated against op_mode.
func TestPortOverrideMirror(t *testing.T) {
	t.Run("RoundTrip", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{"number": 16, "op_mode": "mirror", "mirror_port_idx": 8})
		require.NoError(t, err)
		assert.Equal(t, "mirror", po.OpMode)
		assert.Equal(t, 8, po.MirrorPortIDX)
		assert.Equal(t, 8, fromPortOverride(po)["mirror_port_idx"])

		raw, err := json.Marshal(po)
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"mirror_port_idx":8`)
	})

	t.Run("OmittedWhenNotMirroring", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{"number": 3, "op_mode": "switch"})
		require.NoError(t, err)
		raw, err := json.Marshal(po)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "mirror_port_idx")
	})

	for name, tc := range map[string]struct {
		port, mirror int
		opMode       string
	}{
		"MirrorWithoutSource": {port: 16, opMode: "mirror"},
		"MirrorOfItself":      {port: 16, mirror: 16, opMode: "mirror"},
		"SourceWithoutMirror": {port: 3, mirror: 8, opMode: "switch"},
		"SourceOnDefaultMode": {port: 3, mirror: 8, opMode: ""},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := toPortOverride(map[string]interface{}{"number": tc.port, "op_mode": tc.opMode, "mirror_port_idx": tc.mirror})
			require.Error(t, err)
			assert.Error(t, validatePortMirror(tc.port, tc.opMode, tc.mirror))
		})
	}
}

func TestPortOverrideMirror_SchemaValidator(t *testing.T) {
	elem, ok := ResourceDevice().Schema["port_override"].Elem.(*schema.Resource)
	require.True(t, ok)
	s := elem.Schema["mirror_port_idx"]
	require.NotNil(t, s)
	for _, ok := range []int{0, 1, 56} {
		_, errs := s.ValidateFunc(ok, "mirror_port_idx")
		assert.Empty(t, errs, "%d should be valid", ok)
	}
	for _, bad := range []int{-1, 57} {
		_, errs := s.ValidateFunc(bad, "mirror_port_idx")
		assert.NotEmpty(t, errs, "%d should be invalid", bad)
	}
}

// G23: a non-radio update must not PUT the typed struct's zero-value fields
// (mbb_overrides/config_network/... trip api.err.DeviceNotSupport5gException on
// switches and a blank config_network would wipe the management IP settings).
func TestDeviceUpdateBody(t *testing.T) {
	keys := func(m unifi.DeviceRaw) []string {
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, k)
		}
		return out
	}

	t.Run("SwitchPortChangeSendsOnlyManagedFields", func(t *testing.T) {
		req := &unifi.Device{
			ID:     "dev1",
			MAC:    "70:a7:41:00:00:01",
			SiteID: "default",
			Name:   "USW 16 PoE",
			PortOverrides: []unifi.DevicePortOverrides{
				{PortIDX: 16, Name: "MacBook", Forward: "native", NATiveNetworkID: "net-main", TaggedVLANMgmt: "block_all"},
			},
		}
		body, err := deviceUpdateBody(req, true, false)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"name", "port_overrides"}, keys(body))

		var pos []map[string]interface{}
		require.NoError(t, json.Unmarshal(body["port_overrides"], &pos))
		require.Len(t, pos, 1)
		assert.EqualValues(t, 16, pos[0]["port_idx"])
		assert.Equal(t, "native", pos[0]["forward"])
		assert.Equal(t, "net-main", pos[0]["native_networkconf_id"])
		assert.Equal(t, "block_all", pos[0]["tagged_vlan_mgmt"])
	})

	t.Run("EmptiedOverridesAreSentAsEmptyList", func(t *testing.T) {
		body, err := deviceUpdateBody(&unifi.Device{Name: "sw"}, true, false)
		require.NoError(t, err)
		assert.JSONEq(t, `[]`, string(body["port_overrides"]))
	})

	t.Run("UnchangedEmptyOverridesAreLeftOut", func(t *testing.T) {
		body, err := deviceUpdateBody(&unifi.Device{Name: "ap", BandsteeringMode: "prefer_5g"}, false, false)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"name", "bandsteering_mode"}, keys(body))
	})

	t.Run("EtherLightingOnlyWhenConfigured", func(t *testing.T) {
		req := &unifi.Device{Name: "sw", SwitchVLANEnabled: true, EtherLighting: unifi.DeviceEtherLighting{Mode: "speed"}}
		without, err := deviceUpdateBody(req, false, false)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"name", "switch_vlan_enabled"}, keys(without))

		with, err := deviceUpdateBody(req, false, true)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"name", "switch_vlan_enabled", "ether_lighting"}, keys(with))
	})
}
