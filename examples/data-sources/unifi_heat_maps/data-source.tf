# List all heat maps configured on the site.
data "unifi_heat_maps" "all" {}

output "heat_map_names" {
  value = [for h in data.unifi_heat_maps.all.heat_maps : h.name]
}
