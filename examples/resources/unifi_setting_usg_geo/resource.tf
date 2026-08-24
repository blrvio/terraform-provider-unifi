# Region Blocking (Geo IP filtering) on UniFi Network 10.x.
# Blocks all traffic to/from the listed countries.
resource "unifi_setting_usg_geo" "example" {
  ip_filtering = {
    action            = "block"
    countries         = ["RU", "CN", "KP"]
    traffic_direction = "both"
  }
}
