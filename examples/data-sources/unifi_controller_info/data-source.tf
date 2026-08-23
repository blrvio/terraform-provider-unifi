# Read the controller application info via the Official API (site-independent).
data "unifi_controller_info" "this" {}

output "controller_version" {
  value = data.unifi_controller_info.this.application_version
}
