package network

import "testing"

// TestIntZeroOr covers the "0 (unset) or defer to inner" WAN validators (G3):
// importing a WAN that leaves wan_prefixlen / wan_dhcp_v6_pd_size /
// wireguard_client_peer_port at 0 must not fail validation, while out-of-range
// non-zero values still error.
func TestIntZeroOr(t *testing.T) {
	tests := []struct {
		name      string
		validate  func(interface{}, string) ([]string, []error)
		value     int
		wantError bool
	}{
		{"pd_size zero ok", validateWANDHCPv6PDSize, 0, false},
		{"pd_size in range", validateWANDHCPv6PDSize, 48, false},
		{"pd_size out of range", validateWANDHCPv6PDSize, 47, true},
		{"prefixlen zero ok", validateWANPrefixlen, 0, false},
		{"prefixlen in range", validateWANPrefixlen, 64, false},
		{"prefixlen out of range", validateWANPrefixlen, 129, true},
		{"peer_port zero ok", validateWireguardPeerPortOpt, 0, false},
		{"peer_port valid", validateWireguardPeerPortOpt, 51820, false},
		{"peer_port invalid", validateWireguardPeerPortOpt, 70000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := tt.validate(tt.value, "field")
			if tt.wantError && len(errs) == 0 {
				t.Fatalf("expected an error for value %d, got none", tt.value)
			}
			if !tt.wantError && len(errs) > 0 {
				t.Fatalf("expected no error for value %d, got %v", tt.value, errs)
			}
		})
	}
}
