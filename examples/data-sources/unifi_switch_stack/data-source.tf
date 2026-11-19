# Fetch a single switch stack by its UUID (Official API).
data "unifi_switch_stack" "core" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Each unit is one physical switch of the stack (controller 10.6.109+).
output "core_stack_units" {
  value = {
    for u in data.unifi_switch_stack.core.units : u.id => {
      mac_address = u.mac_address
      role        = u.role # ACTIVE_CONTROLLER, BACKUP_CONTROLLER or MEMBER
    }
  }
}
