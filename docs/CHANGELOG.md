## Changelog for v0.0.60

### Web UI
- The installable version of the web-ui now allows every possible rotation
- Upgraded all third-party `NPM` web dependencies
- Added an option to the system configuration page which sets the `OWM` (*OpenWeatherMap*) Api key
- Added the `material-symbols` dependency for more icon support
- Added system cache deletion button to the system settings page

### Dashboard
- Incremented visual elevation for a better design
- Added automations & schedules as a dashboard component
- Improved dashboard component layout
- Added the current weather as a dashboard component
- Improved display for empty quick actions

### Server
- Removed data races from Go backend tests
- Added efficient weather caching
- Fixed HMS sigTerm termination bug which caused all scripts of a certain ID to be terminated
- Implemented call-stack analysis before dispatching a termination signal
- Fixed security vulnerability which affects the system configuration
- Authentication tokens can now be imported and exported regularly
- Added a `API` route for listing the currently active Homescript jobs
