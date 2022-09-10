## Changelog for v0.2.6

### Bugfixes
#### Web-GUI
- Made the *no quick-actions* text spacing fit to the other dashboard styles
- Added custom behaviour to the automations page when automations are disabled
- Added custom behaviour to the schedules page when schedules are disabled
- Fixed a bug on the profile page which affected the schedules & automations toggle

#### Server & Backend
- When a user's schedules & automations are disabled, schedules and automations will no longer execute
- Any potentially running schedule or automation will be discarded as soon as it would run
