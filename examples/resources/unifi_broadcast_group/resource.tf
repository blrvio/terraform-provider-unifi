# A broadcast group scoping broadcast/multicast traffic to a set of access points.
resource "unifi_broadcast_group" "floor_1" {
  name = "floor-1-aps"
  member_table = [
    "5dc28e5e9106d105bdc87200",
    "5dc28e5e9106d105bdc87201",
  ]
}
