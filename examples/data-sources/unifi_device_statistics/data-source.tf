# Fetch the latest statistics for an adopted device (Official API).
data "unifi_device_statistics" "gateway" {
  device_id = "00000000-0000-0000-0000-000000000000"
}
