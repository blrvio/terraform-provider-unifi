resource "unifi_setting_super_fwupdate" "example" {
  # Release channel for controller (application) updates.
  # Valid options: "internal", "alpha", "beta", "release-candidate", "release".
  controller_channel = "release"

  # Release channel for device firmware updates.
  firmware_channel = "release"

  # Whether Single Sign-On with the Ubiquiti account is enabled for updates.
  sso_enabled = false

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
