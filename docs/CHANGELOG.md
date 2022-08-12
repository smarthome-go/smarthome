## Changelog for v0.0.56

### System Configuration Page
- Added a system configuration page which is accessible under `/system`
- Is able to modify values like
    - The server's geolocation
    - System-wide automation system status
    - System-wide lockdown mode (*currently no functionality*)
- The dashboard includes easy overview and management over the used hardware (*currently only nodes*)
    - View hardware nodes and their online-status
    - Add hardware nodes
    - Edit hardware nodes
    - Delete hardware nodes
- Includes an overview and management over the system's event logs
    - Allows viewing logs (*filtering by log-level*)
    - Allows deletion of old logs (* > 30 days*)
    - Allows deletion of all logs
    - Allows deletion of a single event log entry
- Allows easy configuration `import` and `export` functionality

### Profile Page
- Fixed color bug which occurs when color and the `darkTheme` setting are changed in the same edit
- Added functionality to the manual reload button
    - Reloads user data
    - Reloads permissions
    - Reloads authentication tokens
- Fixed previously broken permission-viewer dialog
- The avatar upload now also allows `UPPERCASE` file endings (*Because e.g. Canon cameras produce `.JPG`*)
- Always reset the token add dialog's input field after a successful submit

### Server Backend
- Improved error log when a node health check fails
- Add logs to node health check
- Addressed trivial API data validation vulnerability
- Added evaluation of system configuration in the setup file runner
- Fixed significant bug in database table deletion function
- The database is now automatically deleted when performing a setup import (*only via the web-UI*)
- If a setup file could not be imported, the `rescue` user with password `rescue` is created
- Removed misleading logs when deleting event logs
- Improved code quality due to refactoring of some database code
- Fixed significant security vulnerability affecting camera feeds
