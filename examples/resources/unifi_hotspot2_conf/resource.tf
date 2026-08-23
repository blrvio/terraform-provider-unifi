# A minimal Passpoint / Hotspot 2.0 configuration profile.
resource "unifi_hotspot2_conf" "example" {
  name = "example-passpoint"

  network_type = 2
  venue_group  = 2
  venue_type   = 8

  network_access_internet = true

  friendly_name = [
    {
      language = "eng"
      text     = "Example Venue"
    },
  ]
}
