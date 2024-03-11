## Changelog for v0.10.1-alpha

### Fixes

- Fixed invalid priority range in notifications
- Added maximum capacity to Homescript runtime output
- Removed debug print from HMS executor
- Fixed various bugs concerning the HMS key-value store
- web: HMS jobs initiated by widgets are now hidden
- The user is no longer informed that a schedule executed
- Prevented multiple consecutive runs of a geolocation-timed automation
- When an HMS error originates from a different file than the executed segment,
  the platform no longer crashes

### Additions

- Automations and schedulers can now be controlled via Homescript
- web: Reminders are now sorted by priority and due-date
- Added Homescript HTTP cookie support
- Added Homescript enumerations
