# Look up an existing spatial record by name.
data "unifi_spatial_record" "floor_1" {
  name = "Floor 1 layout"
}

output "spatial_record_devices" {
  value = data.unifi_spatial_record.floor_1.devices
}
