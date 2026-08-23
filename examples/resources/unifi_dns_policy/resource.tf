# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# An A record.
resource "unifi_dns_policy" "app" {
  type         = "A_RECORD"
  domain       = "app.example.com"
  ipv4_address = "10.0.0.20"
  ttl_seconds  = 300
}

# A DNS forwarding rule.
resource "unifi_dns_policy" "internal_forward" {
  type       = "FORWARD_DOMAIN"
  domain     = "corp.example.com"
  ip_address = "192.168.1.53"
}

# An MX record (disabled).
resource "unifi_dns_policy" "mail" {
  type               = "MX_RECORD"
  domain             = "example.com"
  mail_server_domain = "mail.example.com"
  priority           = 10
  enabled            = false
}
