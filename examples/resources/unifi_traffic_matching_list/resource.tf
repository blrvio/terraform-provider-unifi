# Requires a controller running 10.1.78+ with API-key authentication (Official API).

# A list of IPv4 addresses, ranges and subnets.
resource "unifi_traffic_matching_list" "blocklist" {
  type = "IPV4_ADDRESSES"
  name = "blocklist"

  items = [
    {
      match_type = "IP_ADDRESS"
      value      = "10.0.0.5"
    },
    {
      match_type = "SUBNET"
      value      = "192.168.100.0/24"
    },
    {
      match_type = "IP_ADDRESS_RANGE"
      start      = "10.1.0.1"
      stop       = "10.1.0.100"
    },
  ]
}

# A list of IPv6 addresses and subnets.
resource "unifi_traffic_matching_list" "v6_list" {
  type = "IPV6_ADDRESSES"
  name = "v6-list"

  items = [
    {
      match_type = "IP_ADDRESS"
      value      = "2001:db8::1"
    },
    {
      match_type = "SUBNET"
      value      = "2001:db8::/32"
    },
  ]
}

# A list of ports and port ranges. Provide ports as strings.
resource "unifi_traffic_matching_list" "web_ports" {
  type = "PORTS"
  name = "web-ports"

  items = [
    {
      match_type = "PORT_NUMBER"
      value      = "443"
    },
    {
      match_type = "PORT_NUMBER_RANGE"
      start      = "8000"
      stop       = "8100"
    },
  ]
}
