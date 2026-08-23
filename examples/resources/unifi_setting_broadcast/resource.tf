resource "unifi_setting_broadcast" "example" {
  # Play a sound before a broadcast starts
  sound_before_enabled  = true
  sound_before_resource = "chime"
  sound_before_type     = "sample" # valid values: "sample" or "media"

  # Play a sound after a broadcast ends
  sound_after_enabled  = true
  sound_after_resource = "chime"
  sound_after_type     = "sample" # valid values: "sample" or "media"

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
