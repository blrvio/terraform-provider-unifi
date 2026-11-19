# Requires a controller running 10.1.78+ with API-key authentication (Official API).
data "unifi_vpn_servers" "wireguard" {
  filter = "type.eq('WIREGUARD')"
}

output "wireguard_server_ids" {
  value = [for s in data.unifi_vpn_servers.wireguard.vpn_servers : s.id]
}
