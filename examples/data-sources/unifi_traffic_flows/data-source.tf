# Query observed traffic flows via the go-unifi internal API (API v2).
# All arguments are optional filters; results are returned in `flows`.
data "unifi_traffic_flows" "blocked" {
  action    = ["BLOCK"]
  risk      = ["high"]
  page_size = 50
}

# Narrow by source/destination and a time window (epoch milliseconds).
data "unifi_traffic_flows" "from_host" {
  source_ip      = ["10.0.10.20"]
  timestamp_from = 1735689600000
  timestamp_to   = 1735776000000
}
