# Look up an existing dashboard by name.
data "unifi_dashboard" "ops" {
  name = "Operations"
}

output "dashboard_modules" {
  value = data.unifi_dashboard.ops.modules
}
