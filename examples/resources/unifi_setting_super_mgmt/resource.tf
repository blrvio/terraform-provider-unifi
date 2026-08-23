resource "unifi_setting_super_mgmt" "example" {
  # Automatically upgrade device firmware.
  auto_upgrade = true

  # Device status LEDs.
  led_enabled = true

  # Contact information for the controller / owner.
  contact_info_full_name    = "Ada Lovelace"
  contact_info_company_name = "Analytical Engines"
  contact_info_country      = "US"
  contact_info_city         = "San Francisco"
  contact_info_state        = "CA"
  contact_info_zip          = "94105"

  # Statistics data retention.
  data_retention_setting_preference             = "manual"
  data_retention_time_in_hours_for_hourly_scale = 168
  data_retention_time_in_hours_for_daily_scale  = 720

  # This resource exposes many more controller-wide options (automatic backups,
  # analytics, live updates/chat, SSH credentials, minimum usable disk space,
  # override inform host, etc.). See the resource documentation for the full list.

  # Specify the site (optional, defaults to the site configured in the provider, otherwise "default")
  # site = "default"
}
