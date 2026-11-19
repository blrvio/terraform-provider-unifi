# Fetch a single switch link-aggregation group (LAG) by its UUID (Official API).
data "unifi_switch_lag" "uplink" {
  id = "00000000-0000-0000-0000-000000000000"
}

# members_json follows the LAG variant: deviceId/portIdxs for LOCAL and
# MULTI_CHASSIS, unitId/unitMacAddress/portIdxs for SWITCH_STACK.
output "uplink_members" {
  value = jsondecode(data.unifi_switch_lag.uplink.members_json)
}
