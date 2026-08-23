resource "unifi_setting_mdns" "example" {
  # Reflect only an explicitly chosen set of services across VLANs.
  # Valid options: "all", "auto", "custom"
  mode = "custom"

  # Pick services from the controller's built-in catalog.
  predefined_services {
    code = "apple_airPlay"
  }

  predefined_services {
    code = "google_chromecast"
  }

  # Add an arbitrary DNS-SD service type not covered by the catalog.
  custom_services {
    address = "_ipp._tcp.local"
    name    = "Office Printers"
  }

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
