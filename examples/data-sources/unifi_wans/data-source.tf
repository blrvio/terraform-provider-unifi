# Requires a controller running 10.1.78+ with API-key authentication (Official API).
data "unifi_wans" "all" {}

# Official API UUID of a WAN, looked up by name.
locals {
  wan1_official_id = one([for w in data.unifi_wans.all.wans : w.id if w.name == "Internet 1"])
}
