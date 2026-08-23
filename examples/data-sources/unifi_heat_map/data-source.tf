# Look up an existing heat map by name.
data "unifi_heat_map" "download" {
  name = "Ground floor — download"
}

output "heat_map_type" {
  value = data.unifi_heat_map.download.type
}
