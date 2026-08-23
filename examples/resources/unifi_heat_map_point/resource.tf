resource "unifi_heat_map_point" "example" {
  heatmap_id     = unifi_heat_map.example.id
  x              = 120.5
  y              = 84.0
  download_speed = 480.2
  upload_speed   = 210.6
}
