# A heat map overlay attached to a floor plan (unifi_map), showing download performance.
resource "unifi_map" "ground_floor" {
  name        = "Ground Floor"
  map_type_id = "roadmap"
}

resource "unifi_heat_map" "ground_floor_download" {
  map_id      = unifi_map.ground_floor.id
  name        = "Ground floor — download"
  description = "Measured download throughput"
  type        = "download"
}
