resource "unifi_setting_roaming_assistant" "example" {
  # Enable the Roaming Assistant to nudge clients toward a stronger AP
  enabled = true

  # RSSI (signal strength) threshold in dBm below which clients are
  # encouraged to roam. Valid range: -80 to -60. Only used when enabled.
  rssi = -70

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
