## Changelog for v0.0.48

### Configuration Export
Most of the server's configuration can now be exported using an API endpoint.
The current API endpoint is `/api/config/export` and returns JSON.

Exported settings include:
- Server configuration like `latitude` and `longitude`
- Hardware nodes
- Rooms
    - Room metadata
    - Switches
    - Cameras
- Users
    - User metadata
    - Homescripts
        - Homescript arguments
        - Automations which use Homescripts
    - Reminders
    - Permission settings
    - Switch-permission settings
    - Camera-permission settings

### Better setup.json file
- The concept of the `setup.json` has been improved
- The file now accepts the content of a configuration export, thus allowing Smarthome server migration via the file.
- After the file has been read, it is moved to `./data/config/setup.json.old`
- A setup JSON string can be evaluated via the API (*later via GUI*)

### Homescript
#### Backend (server)
- Added dummy arguments in order for lint will work event when arguments are missing
- During linting, the `getArg` function returns Go's default type values as placeholders
- During linting, the `checkArg` function will always return `true`
- Fixed broken `sigTerm` function (*termination now works as intended*)
- Refactored the HMS manager code (*includes comments and variable names*)

### Frontend (GUI)
- Improved the HMS editor page
    - improved button behavior
    - decreased page title size
    - added more fitting save-state indicator icon
- Improved Homescript deletion confirmation dialog
- Improved automation info dialog
- Fixed bug which causes a cron-expression to be displayed instead of the cron-description
- Refactored Svelte automation component
- Add automatic linting to HMS editor component (using Codemirror's lint-gutter)

### Schedules
#### Backend (server)
- Refactored the internal data structure of the schedule to follow Go's guidelines

#### Frontend (GUI)
- Added a first working version of the scheduler GUI

##### New functionality
- View schedules
- Create schedules
- Edit schedules
- Modify schedules

#### Automation GUI
- Improved the spacing and scaling of the automations GUI

### Miscellaneous
- Refactored and cleaned most of the backend database code
- Fixed some bugs in backend database code
