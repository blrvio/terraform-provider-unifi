# Fetch a single adopted device by its UUID (Official API).
data "unifi_official_device" "gateway" {
  id = "00000000-0000-0000-0000-000000000000"
}
