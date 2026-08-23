resource "unifi_setting_super_sdn" "example" {
  # Whether the UniFi SDN (Ubiquiti cloud account) association is enabled.
  enabled = true

  # Whether the console has been migrated to the cloud account model.
  migrated = true

  # Whether Single Sign-On login is enabled for the SDN association.
  sso_login_enabled = "true"

  # Authentication token used to associate the console with the cloud account.
  # This value is sensitive - prefer supplying it via a variable.
  # auth_token = var.sdn_auth_token

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
