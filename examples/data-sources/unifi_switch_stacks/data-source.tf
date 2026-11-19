# List switch stacks on the site via the Official API.
data "unifi_switch_stacks" "all" {}

# Only the stacks that contain a given unit, filtered server-side.
data "unifi_switch_stacks" "with_unit" {
  filter = "units.macAddress.eq('aa:bb:cc:dd:ee:01')"
}

output "stack_device_ids" {
  value = [for s in data.unifi_switch_stacks.all.stacks : s.device_id]
}
