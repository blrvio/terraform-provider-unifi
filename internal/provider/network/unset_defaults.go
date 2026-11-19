package network

import (
	"github.com/blrvio/go-unifi/v11/unifi"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
)

// Schema defaults the controller stores as "unset" (omitted key: 0 or "") on networks
// that do not use them, such as a VPN server or a WAN.
const (
	defaultNetworkGroup        = "LAN"
	defaultIPV6InterfaceType   = "none"
	defaultDHCPLease           = 86400
	defaultDHCPV6Lease         = 86400
	defaultIPV6RaPreferredLife = 14400
	defaultIPV6RaValidLife     = 86400
)

// applyUnsetDefaults restores the unset fields of current on n after the config overlay.
func applyUnsetDefaults(n, current *unifi.Network) {
	if current == nil {
		return
	}
	n.NetworkGroup = utils.KeepUnset(current.NetworkGroup, n.NetworkGroup, defaultNetworkGroup)
	n.IPV6InterfaceType = utils.KeepUnset(current.IPV6InterfaceType, n.IPV6InterfaceType, defaultIPV6InterfaceType)
	n.DHCPDLeaseTime = utils.KeepUnset(current.DHCPDLeaseTime, n.DHCPDLeaseTime, defaultDHCPLease)
	n.DHCPDV6LeaseTime = utils.KeepUnset(current.DHCPDV6LeaseTime, n.DHCPDV6LeaseTime, defaultDHCPV6Lease)
	n.IPV6RaPreferredLifetime = utils.KeepUnset(current.IPV6RaPreferredLifetime, n.IPV6RaPreferredLifetime, defaultIPV6RaPreferredLife)
	n.IPV6RaValidLifetime = utils.KeepUnset(current.IPV6RaValidLifetime, n.IPV6RaValidLifetime, defaultIPV6RaValidLife)
}
