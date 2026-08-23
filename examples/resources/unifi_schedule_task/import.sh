# import from provider configured site
terraform import unifi_schedule_task.weekly_upgrade 5dc28e5e9106d105bdc87217

# import from another site
terraform import unifi_schedule_task.weekly_upgrade another-site:5dc28e5e9106d105bdc87217
