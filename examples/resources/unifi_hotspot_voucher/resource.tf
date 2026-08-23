# Requires a controller running 10.1.78+ with API-key authentication (Official API).
#
# Vouchers are create-only: the Official API cannot update them, so changing any
# argument forces the voucher to be recreated (generating a new code).

# A simple day-pass voucher (24 hours).
resource "unifi_hotspot_voucher" "day_pass" {
  name               = "front-desk-day-pass"
  time_limit_minutes = 1440
}

# A limited voucher: up to 5 guests, capped data and rate limits.
resource "unifi_hotspot_voucher" "conference" {
  name                    = "conference-2026"
  time_limit_minutes      = 480
  authorized_guest_limit  = 5
  data_usage_limit_mbytes = 2048
  rx_rate_limit_kbps      = 5000
  tx_rate_limit_kbps      = 1000
}

# The generated code is sensitive and populated by the controller after creation.
output "day_pass_code" {
  value     = unifi_hotspot_voucher.day_pass.code
  sensitive = true
}
