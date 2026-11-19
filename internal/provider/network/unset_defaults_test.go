package network

import (
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vpnServer is the shape a UniFi Network 11.0.99 console returns for a WireGuard
// server: no networkgroup, no ipv6_interface_type, zero lease and RA lifetimes (G4/A2).
func vpnServer() *unifi.Network {
	return &unifi.Network{
		ID: "vpn1", Name: "WireGuard Server 1", Purpose: "remote-user-vpn", VPNType: "wireguard-server",
		IPSubnet: "192.168.2.1/24", Enabled: true, InternetAccessEnabled: true,
	}
}

func TestNetworkReadUnsetAsDefault(t *testing.T) {
	t.Parallel()
	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{})
	require.False(t, resourceNetworkSetResourceData(vpnServer(), d, "default").HasError())

	assert.Equal(t, "LAN", d.Get("network_group"))
	assert.Equal(t, "none", d.Get("ipv6_interface_type"))
	assert.Equal(t, 86400, d.Get("dhcp_lease"))
	assert.Equal(t, 86400, d.Get("dhcp_v6_lease"))
	assert.Equal(t, 14400, d.Get("ipv6_ra_preferred_lifetime"))
	assert.Equal(t, 86400, d.Get("ipv6_ra_valid_lifetime"))
}

func TestNetworkWriteKeepsUnset(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		raw          map[string]interface{}
		current      *unifi.Network
		wantGroup    string
		wantV6Type   string
		wantLease    int
		wantRAValid  int
		wantRAPrefer int
	}{
		"VPN with defaults keeps controller unset": {
			raw:     map[string]interface{}{"name": "WireGuard Server 1", "purpose": "remote-user-vpn"},
			current: vpnServer(),
		},
		"explicit non-default is sent": {
			raw: map[string]interface{}{
				"name": "WireGuard Server 1", "purpose": "remote-user-vpn", "dhcp_lease": 3600,
				"ipv6_ra_valid_lifetime": 7200, "ipv6_ra_preferred_lifetime": 3600,
			},
			current: vpnServer(), wantLease: 3600, wantRAValid: 7200, wantRAPrefer: 3600,
		},
		"create sends the defaults": {
			raw:       map[string]interface{}{"name": "lan", "purpose": "corporate"},
			wantGroup: "LAN", wantV6Type: "none", wantLease: 86400, wantRAValid: 86400, wantRAPrefer: 14400,
		},
		"set value on the controller is kept as configured": {
			raw:       map[string]interface{}{"name": "lan", "purpose": "corporate"},
			current:   &unifi.Network{ID: "l", Purpose: "corporate", NetworkGroup: "LAN", IPV6InterfaceType: "none", DHCPDLeaseTime: 86400},
			wantGroup: "LAN", wantV6Type: "none", wantLease: 86400,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, tc.raw)
			n, err := resourceNetworkGetResourceData(d, tc.current)
			require.NoError(t, err)
			assert.Equal(t, tc.wantGroup, n.NetworkGroup)
			assert.Equal(t, tc.wantV6Type, n.IPV6InterfaceType)
			assert.Equal(t, tc.wantLease, n.DHCPDLeaseTime)
			assert.Equal(t, tc.wantRAValid, n.IPV6RaValidLifetime)
			assert.Equal(t, tc.wantRAPrefer, n.IPV6RaPreferredLifetime)
		})
	}
}

// G5: the Official API UUID (networkconf external_id) is exposed as official_id.
func TestNetworkOfficialID(t *testing.T) {
	t.Parallel()
	n := vpnServer()
	n.ExternalID = "663bc70c-85bc-415f-9316-19dbdc2353c0"
	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{})
	require.False(t, resourceNetworkSetResourceData(n, d, "default").HasError())
	assert.Equal(t, "663bc70c-85bc-415f-9316-19dbdc2353c0", d.Get("official_id"))
	assert.True(t, ResourceNetwork().Schema["official_id"].Computed)
	assert.True(t, DataNetwork().Schema["official_id"].Computed)

	// Computed only: the write path keeps the controller's value via read-modify-write.
	out, err := resourceNetworkGetResourceData(schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{"name": "x", "purpose": "remote-user-vpn"}), n)
	require.NoError(t, err)
	assert.Equal(t, n.ExternalID, out.ExternalID)
}
