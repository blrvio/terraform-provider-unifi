resource "unifi_setting_global_ap" "example" {
  # 5 GHz (na) radio defaults
  na_channel_size  = 80  # 20 | 40 | 80 | 160 (MHz)
  na_tx_power_mode = "custom" # auto | medium | high | low | custom
  na_tx_power      = 20 # 0-49 dBm, used when tx_power_mode is "custom"

  # 2.4 GHz (ng) radio defaults
  ng_channel_size  = 20 # 20 | 40 (MHz)
  ng_tx_power_mode = "auto"

  # 6 GHz (6e) radio defaults
  six_e_channel_size  = 160 # 20 | 40 | 80 | 160 (MHz)
  six_e_tx_power_mode = "high"

  # Access points (by MAC) to exclude from these global radio settings.
  ap_exclusions = [
    "aa:bb:cc:dd:ee:ff",
  ]

  # Specify the site (optional, defaults to the site configured in the provider, otherwise "default")
  # site = "default"
}
