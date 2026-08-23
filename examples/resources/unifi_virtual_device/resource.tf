# Place an access-point icon on a floor-plan map.
# A virtual device is a map-placement artifact: it positions an icon on a map,
# it does not configure a live device.
resource "unifi_virtual_device" "lobby_ap" {
  map_id           = "5dc28e5e9106d105bdc87300"
  type             = "uap"
  x                = "120.5"
  y                = "340.0"
  height_in_meters = 2.7
  locked           = true
}
