# Prioritize a set of applications on the primary WAN.
resource "unifi_qos_rule" "critical_apps" {
  name               = "Critical Apps Prioritization"
  enabled            = true
  objective          = "PRIORITIZE"
  download_burst     = "OFF"
  upload_burst       = "OFF"
  wan_or_vpn_network = unifi_network.wan.id

  schedule = {
    mode = "ALWAYS"
  }

  source = {
    matching_target = "ANY"
  }

  destination = {
    matching_target    = "APP"
    port_matching_type = "ANY"
    app_ids            = [393220, 393245]
  }
}
