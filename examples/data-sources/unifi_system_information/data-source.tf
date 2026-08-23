# Read controller-wide system information.
data "unifi_system_information" "this" {}

output "controller_version" {
  value = data.unifi_system_information.this.version
}
