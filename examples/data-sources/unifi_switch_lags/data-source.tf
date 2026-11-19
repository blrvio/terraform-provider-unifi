# List switch link-aggregation groups (LAGs) on the site (Official API).
# The singular unifi_switch_lag data source fetches one by id.
data "unifi_switch_lags" "all" {}

# Device-based LAGs (LOCAL, MULTI_CHASSIS) identify members by device_id;
# SWITCH_STACK LAGs identify them by stack unit (unit_id, unit_mac_address).
output "lag_members" {
  value = {
    for l in data.unifi_switch_lags.all.lags : l.id => [
      for m in l.members : l.type == "SWITCH_STACK" ? "unit ${m.unit_id}: ${join(",", m.port_idxs)}" : "device ${m.device_id}: ${join(",", m.port_idxs)}"
    ]
  }
}
