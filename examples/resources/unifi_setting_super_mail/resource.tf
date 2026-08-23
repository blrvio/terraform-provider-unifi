resource "unifi_setting_super_mail" "example" {
  # Outgoing mail provider used for email notifications.
  # Valid options: "smtp" (custom SMTP server), "cloud" (Ubiquiti relay), or "disabled".
  provider_type = "smtp"

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
