resource "unifi_setting_evaluation_score" "example" {
  # Evaluation score recommendation IDs that have been dismissed.
  dismissed_ids = [
    "ab12",
    "cd345",
  ]

  # Specify the site (optional, defaults to the site configured in the provider, otherwise "default")
  # site = "default"
}
