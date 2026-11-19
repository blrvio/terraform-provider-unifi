data "unifi_dpi_application_categories" "all" {}

# Block every application category for one client, always.
resource "unifi_firewall_app_block" "printer" {
  name        = "[Block:internet] Printer"
  type        = "DEVICE"
  target_type = "APP_CATEGORY"
  client_macs = ["6c:3c:7c:00:00:01"]

  app_category_ids = [for c in data.unifi_dpi_application_categories.all.categories : c.id]

  schedule = {
    mode           = "ALWAYS"
    repeat_on_days = []
    time_all_day   = false
  }
}
