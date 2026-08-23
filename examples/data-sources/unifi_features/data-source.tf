# List all controller features and their availability on the site.
data "unifi_features" "all" {}

output "available_features" {
  value = [for f in data.unifi_features.all.features : f.name if f.feature_exists]
}
