resource "unifi_setting_radio_ai" "example" {
  # Enable AI-driven radio optimization ("AI WiFi").
  enabled = true

  # Manage the parameters explicitly rather than leaving everything on auto.
  # Valid options: "auto", "manual"
  setting_preference = "manual"

  # Let the optimizer adjust both channel and transmit power.
  optimize = ["channel", "power"]

  # Optimize the 5 GHz and 2.4 GHz radios.
  radios = ["na", "ng"]

  # Run the optimization nightly at 03:00.
  cron_expr = "0 3 * * *"

  # Constrain eligible channels to the country's regulatory domain.
  auto_adjust_channels_to_country = true

  # Restrict the 5 GHz radio to a known-good channel set and widths.
  channels_na = [36, 40, 44, 48, 149, 153, 157, 161]
  ht_modes_na = [20, 40, 80]

  # Never let the optimizer pick this channel/width on 5 GHz.
  channels_blacklist {
    channel       = 144
    channel_width = 160
    radio         = "na"
  }

  # Per-radio tuning: allow 80 MHz on 5 GHz but keep DFS channels off.
  radios_configuration {
    radio         = "na"
    channel_width = 80
    dfs           = false
  }

  # Keep a critical AP out of the optimization entirely.
  # exclude_devices = ["aa:bb:cc:dd:ee:ff"]

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
