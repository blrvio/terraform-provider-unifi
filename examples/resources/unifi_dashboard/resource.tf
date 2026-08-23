# A custom dashboard composed of modules.
resource "unifi_dashboard" "ops" {
  name      = "Operations"
  desc      = "Key operational widgets"
  is_public = false

  modules = [
    {
      id           = "module-1"
      module_id    = "traffic-overview"
      config       = jsonencode({ range = "24h" })
      restrictions = ""
    },
  ]
}
