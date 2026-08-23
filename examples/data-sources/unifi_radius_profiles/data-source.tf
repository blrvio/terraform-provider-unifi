# List RADIUS profiles on the site via the Official API (read-only).
# Named distinctly from the internal unifi_radius_profile resource.
data "unifi_radius_profiles" "all" {}
