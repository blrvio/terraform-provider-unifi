resource "unifi_setting_super_events" "example" {
  # This console-level setting exposes only a single opaque field, surfaced
  # here as "ignored" (the controller's "_ignored" field). Typically left
  # unset and managed by the controller.
  # ignored = ""

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
