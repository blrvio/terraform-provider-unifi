# import from provider configured site
terraform import unifi_qos_rule.critical_apps 694009322cd2eb1c05176d01

# import from another site
terraform import unifi_qos_rule.critical_apps another-site:694009322cd2eb1c05176d01
