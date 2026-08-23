# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# A standard WPA2-Personal WiFi broadcast (SSID).
resource "unifi_wifi_broadcast" "corp" {
  name = "Corp-WiFi"
  type = "STANDARD"

  # security_configuration accepts the Official API object shape as JSON.
  security_configuration = jsonencode({
    type       = "WPA2_PERSONAL"
    passphrase = "change-me-please"
  })
}

# An IoT-optimized, hidden broadcast with client (MAC) filtering.
resource "unifi_wifi_broadcast" "iot" {
  name                     = "IoT-WiFi"
  type                     = "IOT_OPTIMIZED"
  hide_name                = true
  client_isolation_enabled = true

  security_configuration = jsonencode({
    type       = "WPA2_PERSONAL"
    passphrase = "another-secret"
  })

  # Optional nested objects also accept the Official API object shape as JSON.
  client_filtering_policy = jsonencode({
    action           = "BLOCK"
    macAddressFilter = ["aa:bb:cc:dd:ee:ff"]
  })
}
