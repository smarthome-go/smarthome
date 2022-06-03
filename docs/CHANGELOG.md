## Changelog for v0.0.39

### Homescript
- Removed legacy *RadiGo* features due to the recent addition of network functions
- Added a first working version of a Homescript management web GUI

### Rooms
- Improved camera-feed scaling: added better widescreen support to the `viewCamera` dialog

### Automations
- Fixed 1 bug in automation creation: selected days are now also reset on addition and cancelation

### Backend
- Significantly improved power-request response times when a node is offline
- Increased backend reliabilty through the use of URL builders instead of string formatting

### Development
- Added the `IconPicker.svelte` component: It can be used to allow the selection of any compatible Google-MD icon from a given icon set
