# Look up an existing tag by name.
data "unifi_tag" "iot" {
  name = "iot-devices"
}

output "iot_tag_members" {
  value = data.unifi_tag.iot.member_table
}
