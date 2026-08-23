# List switch link-aggregation groups (LAGs) on the site (Official API).
# The singular unifi_switch_lag data source fetches one by id.
data "unifi_switch_lags" "all" {}
