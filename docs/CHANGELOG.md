## Changelog for v0.1.0

*Note*: This is a significant release because it marks the official end of active, breaking development.  
Changes from here will follow the `SemVer` guidelines and will try to reduce breaking changes as much as possible.

### Web-UI
- Removed debug-logs from the web-UI
- Removed the `material-symbols` dependency due to irrational network-loads
- Automatically detect nighttime and use weather icon accordingly
- Cleaned the Web-UI code and removed unneeded parts

### Homescript
- Add workspaces which allow visual groups of coherent scripts
- The URL of the HMS-editor is now changed reactively which allows sharing the URL
- The HMS-execution via the `F-X` keys is now prevented when a script is already running
- Improved the HMS editor's mobile layout

### Server
- The sunrise & sunset time is now included in every weather response
