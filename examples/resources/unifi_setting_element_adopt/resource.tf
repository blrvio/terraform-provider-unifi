resource "unifi_setting_element_adopt" "example" {
  # Enable adoption of UniFi Elements devices
  enabled = true

  # Dedicated SSID and pre-shared key used for Element adoption.
  # Only used when enabled. The PSK is sensitive.
  x_element_essid = "element-adopt"
  x_element_psk   = "change-me-please"

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
