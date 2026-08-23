# List all dashboards configured on the site.
data "unifi_dashboards" "all" {}

output "dashboard_names" {
  value = [for d in data.unifi_dashboards.all.dashboards : d.name]
}
