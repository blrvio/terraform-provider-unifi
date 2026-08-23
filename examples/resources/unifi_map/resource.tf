resource "unifi_map" "example" {
  name        = "ground-floor"
  type        = "imageMap"
  unit        = "m"
  opacity     = 0.8
  offset_left = 0
  offset_top  = 0
  upp         = 1.0
  zoom        = 18
  selected    = true
}
