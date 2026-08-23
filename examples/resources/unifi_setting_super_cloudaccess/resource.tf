resource "unifi_setting_super_cloudaccess" "example" {
  # Whether remote cloud access is enabled for the console.
  enabled = true

  # The AWS IoT certificate material below is only applicable when enabled = true.
  # These values are sensitive - prefer supplying them via variables.
  # device_auth       = var.cloudaccess_device_auth
  # x_certificate_arn = var.cloudaccess_certificate_arn
  # x_certificate_pem = var.cloudaccess_certificate_pem
  # x_private_key     = var.cloudaccess_private_key

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
