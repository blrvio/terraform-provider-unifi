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
