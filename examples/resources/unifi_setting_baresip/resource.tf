resource "unifi_setting_baresip" "example" {
  # Enable the Baresip SIP client
  enabled = true

  # SIP server to register with, and optional outbound proxy.
  # Only used when enabled.
  server         = "sip.example.com"
  outbound_proxy = "proxy.example.com"

  # URL of the Baresip package to install (must be a valid URL)
  package_url = "https://example.com/packages/baresip.pkg"

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
