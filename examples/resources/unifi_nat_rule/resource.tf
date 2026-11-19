# Experimental. Forward TCP 8443 on the WAN to 10.0.10.20:443.
resource "unifi_nat_rule" "https_to_server" {
  description  = "https to server"
  type         = "DNAT"
  in_interface = "<wan networkconf id>"
  protocol     = "tcp"
  ip_address   = "10.0.10.20"
  port         = "443"

  destination_filter = {
    filter_type = "ADDRESS_AND_PORT"
    port        = "8443"
  }
}
