resource "unifi_hotspot_package" "day_pass" {
  name       = "Day Pass"
  amount     = 9.99
  currency   = "USD"
  charged_as = "per day"
  hours      = 24

  limit_overwrite = true
  limit_down      = 5000
  limit_up        = 2000
  limit_quota     = 1024

  payment_fields_email_enabled  = true
  payment_fields_email_required = true
}
