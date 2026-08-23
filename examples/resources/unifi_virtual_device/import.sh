# import from provider configured site
terraform import unifi_virtual_device.lobby_ap 5dc28e5e9106d105bdc87217

# import from another site
terraform import unifi_virtual_device.lobby_ap another-site:5dc28e5e9106d105bdc87217
