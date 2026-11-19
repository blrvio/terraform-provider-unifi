# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# Turn on syslog logging for a predefined policy (e.g. "Block Invalid Traffic").
resource "unifi_official_firewall_predefined_policy" "block_invalid" {
  policy_id       = "0b6a8c0e-0000-4000-8000-000000000001"
  logging_enabled = true
}
