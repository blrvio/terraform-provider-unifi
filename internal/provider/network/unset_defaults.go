package network

import "github.com/blrvio/go-unifi/v11/unifi"

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

// orDefault reports an unset controller value as the schema default, so a config that
// omits the attribute (and so takes the default) matches an imported network.
func orDefault[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

// keepUnset returns the zero value when the controller had the field unset and the
// config still asks for the schema default, so an update never writes defaults the
// controller did not have (e.g. network_group "LAN" onto a VPN server).
func keepUnset[T comparable](current, planned, def T) T {
	var zero T
	if current == zero && planned == def {
		return zero
	}
	return planned
}

// applyUnsetDefaults restores the unset fields of current on n after the config overlay.
func applyUnsetDefaults(n, current *unifi.Network) {
	if current == nil {
		return
	}
	n.NetworkGroup = keepUnset(current.NetworkGroup, n.NetworkGroup, defaultNetworkGroup)
	n.IPV6InterfaceType = keepUnset(current.IPV6InterfaceType, n.IPV6InterfaceType, defaultIPV6InterfaceType)
	n.DHCPDLeaseTime = keepUnset(current.DHCPDLeaseTime, n.DHCPDLeaseTime, defaultDHCPLease)
	n.DHCPDV6LeaseTime = keepUnset(current.DHCPDV6LeaseTime, n.DHCPDV6LeaseTime, defaultDHCPV6Lease)
	n.IPV6RaPreferredLifetime = keepUnset(current.IPV6RaPreferredLifetime, n.IPV6RaPreferredLifetime, defaultIPV6RaPreferredLife)
	n.IPV6RaValidLifetime = keepUnset(current.IPV6RaValidLifetime, n.IPV6RaValidLifetime, defaultIPV6RaValidLife)
}
