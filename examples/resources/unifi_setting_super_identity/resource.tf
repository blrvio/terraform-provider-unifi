resource "unifi_setting_super_identity" "example" {
  # Network hostname of the UniFi console (must be a valid hostname)
  hostname = "udm-pro"

  # Human-readable display name shown in the UniFi interface
  name = "Head Office"

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
