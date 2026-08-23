# import from provider configured site
terraform import unifi_dhcp_option.captive_portal_url 5dc28e5e9106d105bdc87217

# import from another site
terraform import unifi_dhcp_option.captive_portal_url another-site:5dc28e5e9106d105bdc87217
