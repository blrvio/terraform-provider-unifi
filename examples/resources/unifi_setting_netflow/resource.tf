resource "unifi_setting_netflow" "example" {
  # Enable NetFlow export to an external collector.
  # When disabled, none of the fields below may be set.
  enabled = true

  # NetFlow collector endpoint.
  server = "10.0.0.5"
  port   = 2055 # 1-65535

  # Protocol version: 5, 9, or 10.
  version = 9

  # Export and template refresh cadence (seconds).
  export_frequency = 60
  refresh_rate     = 30

  # Flow sampling: off | hash | random | deterministic
  sampling_mode = "hash"
  sampling_rate = 100

  # Engine identification.
  auto_engine_id_enabled = true

  # Networks (by ID) to export flow records for.
  network_ids = [
    unifi_network.lan.id,
  ]

  # Specify the site (optional, defaults to the site configured in the provider, otherwise "default")
  # site = "default"
}
