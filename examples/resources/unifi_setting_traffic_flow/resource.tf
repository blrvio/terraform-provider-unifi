resource "unifi_setting_traffic_flow" "example" {
  # Control which categories of traffic the controller may observe and manage
  enabled_allowed_traffic         = true
  gateway_dns_enabled             = true
  unifi_device_management_enabled = true
  unifi_services_enabled          = false

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
