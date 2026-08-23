# A hotspot operator account used to authorise guest hotspot access.
resource "unifi_hotspot_operator" "front_desk" {
  name       = "front-desk"
  note       = "Lobby reception operator"
  x_password = "change-me"
}
