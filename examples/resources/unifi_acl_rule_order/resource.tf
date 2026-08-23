# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# Manage the priority order of ACL rules for a site. The first id has the
# highest priority. This is a singleton resource — one per site.
resource "unifi_acl_rule_order" "main" {
  rule_ids = [
    unifi_acl_rule.block_guests.id,
    unifi_acl_rule.allow_printer.id,
  ]
}
