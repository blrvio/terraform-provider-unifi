package network

import (
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// R1: a unifi_wlan update starts from the controller object; unmodeled fields and the
// UI-chosen WPA mode / DTIM / rekey survive, managed attributes are overlaid.
func TestWLANGetResourceDataReadModifyWrite(t *testing.T) {
	t.Parallel()
	current := &unifi.WLAN{
		ID: "w1", SiteID: "139a62c9", Name: "old", Security: "wpapsk",
		WPAMode: "wpa2", WPAEnc: "gcmp", GroupRekey: 0, DTIMMode: "custom", DTIMNa: 3,
		Enabled: false, NameCombineEnabled: false, IappEnabled: true, BSupported: true,
	}
	raw := map[string]interface{}{"name": "new", "security": "wpapsk", "passphrase": "supersecret"}
	d := schema.TestResourceDataRaw(t, ResourceWLAN().Schema, raw)
	meta := &base.Client{Site: "default", Version: version.Must(version.NewVersion("11.0.99"))}

	got, err := resourceWLANGetResourceData(d, meta, current)
	require.NoError(t, err)
	assert.Equal(t, "new", got.Name, "managed attribute overlaid")
	assert.Equal(t, "supersecret", got.XPassphrase)
	assert.Equal(t, "gcmp", got.WPAEnc, "create-time default must not overwrite the controller value")
	assert.Equal(t, 0, got.GroupRekey)
	assert.Equal(t, "custom", got.DTIMMode)
	assert.Equal(t, 3, got.DTIMNa, "unmodeled field kept")
	assert.False(t, got.Enabled, "a WLAN disabled in the UI is not re-enabled by an unrelated update")
	assert.True(t, got.IappEnabled)
	assert.True(t, got.BSupported)
	assert.Equal(t, "139a62c9", got.SiteID)

	created, err := resourceWLANGetResourceData(d, meta, nil)
	require.NoError(t, err)
	assert.Equal(t, "ccmp", created.WPAEnc, "create still applies the defaults")
	assert.True(t, created.Enabled)
	assert.Equal(t, 3600, created.GroupRekey)
}

// G15: the controller returns wlan_band and wlan_bands together; state (and generated
// config) must carry only one, since the attributes conflict.
func TestWLANBandAttributes(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		resp      unifi.WLAN
		prior     map[string]interface{}
		wantBand  interface{}
		wantBands bool
	}{
		"import auto both":     {resp: unifi.WLAN{WLANBand: "both", WLANBands: []string{"2g", "5g"}, SettingPreference: "auto"}, wantBand: "both"},
		"import manual":        {resp: unifi.WLAN{WLANBand: "both", WLANBands: []string{"2g", "5g"}, SettingPreference: "manual"}, wantBands: true},
		"import 6g":            {resp: unifi.WLAN{WLANBand: "both", WLANBands: []string{"2g", "5g", "6g"}}, wantBands: true},
		"state uses wlan_band": {resp: unifi.WLAN{WLANBand: "5g", WLANBands: []string{"5g"}, SettingPreference: "manual"}, prior: map[string]interface{}{"wlan_band": "5g"}, wantBand: "5g"},
		"state uses wlan_bands": {resp: unifi.WLAN{WLANBand: "both", WLANBands: []string{"2g", "5g"}, SettingPreference: "auto"},
			prior: map[string]interface{}{"wlan_bands": []interface{}{"2g", "5g"}}, wantBands: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			raw := tc.prior
			if raw == nil {
				raw = map[string]interface{}{}
			}
			d := schema.TestResourceDataRaw(t, ResourceWLAN().Schema, raw)
			band, bands := wlanBandAttributes(&tc.resp, d)
			if tc.wantBands {
				assert.Nil(t, band)
				assert.NotNil(t, bands)
			} else {
				assert.Equal(t, tc.wantBand, band)
				assert.Nil(t, bands)
			}
		})
	}
}
