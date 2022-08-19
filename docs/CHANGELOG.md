## Changelog for v0.0.60

### Homescript Web-UI
- Added a visual indicator for running scripts to the HMS list selector

### Dashboard Web-UI
- Allowed the possibility to terminate scripts which are already running from the *Quick Actions* component on the dashboard
- Optimized the *Quick Actions* display for mobile devices
- Finalized the *Quick Actions* alignment trough invisible dummies
- Tweaked the general reactivity of the dashboard, for example the *power usage chart*'s scaling
- Added a *reminder* dashboard component for viewing and removing reminders
- Added some visual padding to the dashboard's weather component's contents
- Fixed failing no-automations detection in dashboard *automations / schedules* component

### Reminder Web-UI
- Fixed the misleading edit-button behavior which was occurring when modifying a reminder
- The reminders on the reminder page are now sorted by-priority (*in descending order*)

### General Web-UI
- Added a much more human-friendly way of displaying a notification's time
- Made the *nav bar* much more mobile-friendly (*especially on landscape orientation*)

### Server
- Removed undesired debug print statements from backend code
- Removed unneeded *Smarthome-CLI* from the server's docker-image
- This decreased the docker-image size by `20.7%` in comparison to `v0.0.59`
- Added functionality to the *lockdown mode*, which will now block any form of switch-request
