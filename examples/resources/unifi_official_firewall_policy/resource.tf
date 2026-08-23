# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# The source, destination, ip_protocol_scope and schedule attributes are
# JSON-encoded strings matching the Official-API object shapes.
resource "unifi_official_firewall_policy" "allow_iot_to_wan" {
  name   = "allow-iot-to-wan"
  action = "ALLOW"

  source = jsonencode({
    zoneId = "11111111-1111-1111-1111-111111111111"
  })
  destination = jsonencode({
    zoneId = "22222222-2222-2222-2222-222222222222"
  })
  ip_protocol_scope = jsonencode({
    ipVersion = "BOTH"
    type      = "ALL"
  })
}

# A logging BLOCK policy that only matches new/invalid connections.
resource "unifi_official_firewall_policy" "block_invalid" {
  name                    = "block-invalid"
  action                  = "BLOCK"
  logging_enabled         = true
  description             = "Drop new and invalid connections."
  connection_state_filter = ["NEW", "INVALID"]

  source = jsonencode({
    zoneId = "11111111-1111-1111-1111-111111111111"
  })
  destination = jsonencode({
    zoneId = "22222222-2222-2222-2222-222222222222"
  })
  ip_protocol_scope = jsonencode({
    ipVersion = "BOTH"
    type      = "ALL"
  })
}
