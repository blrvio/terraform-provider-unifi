# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# A firewall zone grouping one or more networks by their UUIDs.
resource "unifi_official_firewall_zone" "iot" {
  name = "IoT"
  network_ids = [
    "44444444-4444-4444-4444-444444444444",
    "55555555-5555-5555-5555-555555555555",
  ]
}
