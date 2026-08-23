# Look up a single connected client by its UUID (Official API).
# This surface has no MAC field, so lookup is by client UUID only.
data "unifi_client" "one" {
  id = "00000000-0000-0000-0000-000000000000"
}
