# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# Order the user-defined firewall policies for a source/destination zone pair.
resource "unifi_official_firewall_policy_order" "iot_to_wan" {
  source_zone_id      = "11111111-1111-1111-1111-111111111111"
  destination_zone_id = "22222222-2222-2222-2222-222222222222"
  policy_ids = [
    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
  ]
}
