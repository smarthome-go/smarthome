## Changelog for 0.0.32

### Rooms Frontend
- (*Bugfix*) Fixed wrong alignment of add room button in rooms GUI
### Backend
- (*Addition*) Added `SMARTHOME_ENV_PRODUCTION` environment variable to the server
  - During docker-image build, the variable is set to `true` by default
  - Is now included in the `docker-compose.yml` as example
- (*Improvement*) Modified startup log messages to be more precise
- (*Bugfix*) Fixed bad user-color bug: added default user-colors to homescript `AddUser` function

### Development
- (*Improvement*) Restructured development docker directory
  - (*Bugfix*) Prevented interference of `foo-bar-docker-compose.yml` files in docker directory