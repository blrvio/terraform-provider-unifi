# Report the resources referencing a network (clients, devices, WiFi, ...).
data "unifi_network_references" "lan" {
  network_id = "00000000-0000-0000-0000-000000000000"
}
