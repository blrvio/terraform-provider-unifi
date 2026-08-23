# A spatial record placing access points at x/y/z coordinates within a mapped space.
resource "unifi_spatial_record" "floor_1" {
  name = "Floor 1 layout"

  devices = [
    {
      mac = "aa:bb:cc:dd:ee:ff"
      position = {
        x = 12.5
        y = 4.0
        z = 3.0
      }
    },
  ]
}
