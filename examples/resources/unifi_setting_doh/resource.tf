resource "unifi_setting_doh" "example" {
  # Encrypt outbound DNS using custom DNS-over-HTTPS resolvers.
  # Valid options: "off", "auto", "manual", "custom"
  state = "custom"

  # When state = "manual", pick from well-known providers instead:
  # server_names = ["cloudflare", "google"]

  # Define a custom resolver via its DNS Stamp (sdns://...).
  custom_servers {
    enabled     = true
    server_name = "cloudflare-secure"
    sdns_stamp  = "sdns://AgcAAAAAAAAAAAAHOS45LjkuOQ"
  }

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
