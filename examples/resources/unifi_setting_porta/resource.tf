resource "unifi_setting_porta" "example" {
  # Enable the WAN2 port on a UniFi Security Gateway 3P (USG-3P)
  ugw3_wan2_enabled = true

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
