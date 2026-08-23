# A weekly scheduled firmware upgrade for two devices, at 04:00 every Sunday.
resource "unifi_schedule_task" "weekly_upgrade" {
  name      = "weekly-upgrade"
  action    = "upgrade"
  cron_expr = "0 4 * * 0"
  upgrade_targets = [
    "aa:bb:cc:dd:ee:ff",
    "11:22:33:44:55:66",
  ]
}

# A one-time upgrade that removes itself after running.
resource "unifi_schedule_task" "one_off_upgrade" {
  name              = "maintenance-window"
  cron_expr         = "30 2 15 6 *"
  execute_only_once = true
  upgrade_targets   = ["aa:bb:cc:dd:ee:ff"]
}
