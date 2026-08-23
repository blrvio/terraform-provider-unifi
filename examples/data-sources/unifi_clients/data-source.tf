# List all clients currently connected to the default site (Official API).
data "unifi_clients" "all" {}

# Restrict to a single site and apply an optional server-side filter.
data "unifi_clients" "guests" {
  site   = "default"
  filter = "type.eq('WIRELESS')"
}
