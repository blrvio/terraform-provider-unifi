resource "unifi_setting_snmp" "example" {
  # Enable the SNMP v1/v2c agent and expose metrics under a community string.
  enabled   = true
  community = "public"

  # Optionally enable the more secure SNMP v3 agent with user/password auth.
  enabled_v3 = true
  username   = "monitor"
  x_password = "change-me-please" # sensitive

  # Specify the site (optional, defaults to the provider's site, otherwise "default")
  # site = "default"
}
