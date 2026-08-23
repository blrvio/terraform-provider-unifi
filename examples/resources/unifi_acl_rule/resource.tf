# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# An IP-based ACL rule blocking a source network.
# The *_filter attributes are polymorphic JSON — supply them with jsonencode().
resource "unifi_acl_rule" "block_guests" {
  name   = "block-guests"
  action = "BLOCK"
  type   = "IPV4"

  source_filter = jsonencode({
    type      = "NETWORK"
    networkId = "11111111-2222-3333-4444-555555555555"
  })
}

# A MAC-based ACL rule (disabled) with a description.
resource "unifi_acl_rule" "allow_printer" {
  name        = "allow-printer"
  action      = "ALLOW"
  type        = "MAC"
  enabled     = false
  description = "Temporarily disabled"

  source_filter = jsonencode({
    type = "MAC"
    mac  = "00:11:22:33:44:55"
  })
}

# Ordering (priority) is managed separately — see unifi_acl_rule_order.
