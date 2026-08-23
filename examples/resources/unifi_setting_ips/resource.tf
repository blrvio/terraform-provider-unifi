
resource "unifi_network" "test" {
  name = "My Network"
  purpose = "corporate"
  subnet = "192.168.1.0/24"
  vlan_id = 10
}

resource "unifi_setting_ips" "example" {
  # Set IPS mode to "ips" (Intrusion Prevention System)
  # Other valid options: "ids" (Intrusion Detection System) or "disabled"
  ips_mode = "ips"
  
  # Networks on which IPS/IDS should be enabled
  enabled_networks = [unifi_network.test.id]
  
  # Advanced filtering preference
  # Valid options: "disabled", "manual", or "auto"
  advanced_filtering_preference = "manual"
  
  # Categories of threats to detect/prevent
  enabled_categories = [
    "emerging-dos",
    "emerging-exploit",
    "emerging-malware"
  ]
  
  # Honeypot configuration
  honeypots = [
    {
      ip_address = "192.168.1.10"
      network_id = unifi_network.test.id
    }
  ]

  # Note: ad blocking and per-network DNS filtering were removed from the IPS
  # setting in controller v10+ (go-unifi v10). Manage those via DNS policies /
  # content filtering instead.

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
