# Failover: WAN is primary, WAN2 takes over when it goes down.
resource "unifi_wan_load_balancing" "this" {
  mode = "FAILOVER_ONLY"

  wan_interfaces = {
    WAN  = { mode = "DISTRIBUTED", priority = 1, weight = 50 }
    WAN2 = { mode = "FAILOVER_ONLY", priority = 2, weight = 50 }
  }
}
