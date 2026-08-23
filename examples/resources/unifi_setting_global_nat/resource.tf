resource "unifi_setting_global_nat" "example" {
  # Global NAT mode applied site-wide.
  # Valid options: "auto", "custom", "off"
  mode = "custom"

  # Networks excluded from global NAT (by network ID).
  excluded_network_ids = [
    unifi_network.iot.id,
  ]

  # Specify the site (optional, defaults to the site configured in the provider, otherwise "default")
  # site = "default"
}
