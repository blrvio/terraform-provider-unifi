# import from provider configured site
terraform import unifi_firewall_zone_policy.mypolicy 5dc28e5e9106d105bdc87217

# import from another site
terraform import unifi_firewall_zone_policy.mypolicy another-site:5dc28e5e9106d105bdc87217
