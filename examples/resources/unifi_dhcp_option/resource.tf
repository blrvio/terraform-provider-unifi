# A custom DHCP option (option 114, the captive-portal URL) served as text.
resource "unifi_dhcp_option" "captive_portal_url" {
  code = "114"
  name = "captive-portal"
  type = "text"
}

# A numeric option with an explicit signed 16-bit width.
resource "unifi_dhcp_option" "lease_hint" {
  code   = "150"
  name   = "lease-hint"
  type   = "integer"
  width  = 16
  signed = false
}
