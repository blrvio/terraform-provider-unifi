resource "unifi_setting_super_smtp" "example" {
  # Whether the custom SMTP server is enabled.
  enabled = true

  # The SMTP server attributes below are only applicable when enabled = true.
  host = "smtp.example.com"
  port = 587

  # Use a custom sender (From) address for outgoing messages.
  use_sender = true
  sender     = "noreply@example.com"

  # Use SSL/TLS when connecting to the SMTP server.
  use_ssl = true

  # Authenticate to the SMTP server. The credentials are sensitive -
  # prefer supplying them via variables.
  use_auth = true
  username = "smtp-user"
  # x_password = var.smtp_password

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
