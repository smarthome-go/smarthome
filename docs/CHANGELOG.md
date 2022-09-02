## Changelog for v0.2.3

### Bugfixes
#### Web-GUI
- The Homescript toggle settings are now marginally smaller, which allows a better display on mobile devices
- Renamed the `Automations` toggle of Homescripts to `Show Selection` for a more concise description
- Added a helper text which describes why there are no Homescripts in the scheduler target selection
- Fixed misalignment of the reload button at the top of the scheduler page
- Added a missing header title to the Homescript manager page
- Cleaned up some potential buggy code in the Homescript manager page
- Temporarily pinned the version a broken UI library component

#### Server & Backend
- Added a missing `sysctl` entry to the exampl `docker-compose.yml` file
- This allows the Homescript `ping` function to be used inside a Smarthome Docker container
- Fixed a bad log formatting in the server
