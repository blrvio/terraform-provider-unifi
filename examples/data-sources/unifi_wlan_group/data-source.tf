# Look up the default WLAN group.
data "unifi_wlan_group" "default" {
}

# Look up a WLAN group by name.
data "unifi_wlan_group" "guest" {
  name = "guest-group"
}
