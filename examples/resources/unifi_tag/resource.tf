# A tag grouping a set of member identifiers (e.g. device MAC addresses).
resource "unifi_tag" "iot" {
  name = "iot-devices"
  member_table = [
    "aa:bb:cc:dd:ee:ff",
    "11:22:33:44:55:66",
  ]
}

# A tag with no members yet.
resource "unifi_tag" "printers" {
  name = "printers"
}
