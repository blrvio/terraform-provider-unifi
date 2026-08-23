# Check whether a named controller feature is available on the site.
data "unifi_feature" "zone_based_firewall" {
  name = "ZONE_BASED_FIREWALL"
}

output "zbf_available" {
  value = data.unifi_feature.zone_based_firewall.feature_exists
}
