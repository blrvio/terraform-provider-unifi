# A content filtering rule for kids' devices: family + advertisement categories,
# safe search on Google/YouTube, active on weekday evenings.
resource "unifi_content_filtering" "kids" {
  name    = "kids"
  enabled = true

  categories  = ["FAMILY", "ADVERTISEMENT"]
  safe_search = ["GOOGLE", "YOUTUBE"]

  block_list = ["bad.example"]
  allow_list = ["school.example.edu"]

  client_macs = ["aa:bb:cc:dd:ee:ff"]

  schedule = {
    mode             = "EVERY_WEEK"
    repeat_on_days   = ["mon", "tue", "wed", "thu", "fri"]
    time_range_start = "16:00"
    time_range_end   = "21:00"
  }
}

# A simple always-on rule applied to a whole network.
resource "unifi_content_filtering" "guest_network" {
  name        = "guest-filter"
  categories  = ["ADVERTISEMENT"]
  network_ids = ["5dc28e5e9106d105bdc87217"]
}
