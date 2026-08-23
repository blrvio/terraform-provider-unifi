resource "unifi_setting_dashboard" "example" {
  # Curate the dashboard cards manually instead of letting the controller decide.
  # Valid options: "auto", "manual"
  layout_preference = "manual"

  # Show the WAN activity card.
  widgets {
    name    = "wan_activity"
    enabled = true
  }

  # Hide the "most active apps" card.
  widgets {
    name    = "most_active_apps"
    enabled = false
  }

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
