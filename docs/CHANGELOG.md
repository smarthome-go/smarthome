## Changelog for 0.0.30

### Rooms Page
- Added a *local settings* dialog for configuring the reload behavior of cameras for the current device
- Fixed switch-modification in GUI via intermediate stage
- Optimized GUI code for accessibility and better scaling
- Improved display of cam GUI (add perm. indicator)
- Added additional texts for empty sections in rooms
- Accounted for cameras when listing personal rooms
- Added camera-permissions frontend in user settings
- Added better display for when the user has no rooms

### User Management
- Changed user-permissions editor icon
### Backend
- Added camera permissions API backend
- Accounted for cameras when listing personal rooms
- (*Bugfix*) Added safety checks for user management backend
- (*Bugfix*) Added safety validation when deleting a Homescript 
- (*Bugfix*) Removed double-log in camera backend error
- (*Bugfix*) Fixed additional 2 bugs in backend camera-permissions
- (*Improvement*) Improved efficiency of ensureValidFormat in cams

### Workflow
- Added Github-CLI makefile target for creating a GH release