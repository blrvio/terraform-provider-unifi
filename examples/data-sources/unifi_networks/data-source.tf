# List the networks on the site via the Official API (read-only).
# Writable network management remains the internal unifi_network resource.
data "unifi_networks" "all" {}
