## Changelog for v0.0.47

### Homescript
- Added a global *manager* to the Homescript module
- The manager orchestrates the way Homescripts are executed and terminated
- As of Homescript release `v0.13.0`, sigTerm can be used to terminate a script's execution at any point in time
- The API routes `/api/homescript/kill/script/{id}` and `/api/homescript/kill/job/{id}` can now be used to terminate Homescript jobs.
- Added a visual *cancel* button to the Homescript editor GUI which terminates all jobs running the currently selected script
- Added minor tweaks to the HMS editor GUI which increase stability and overall consistency
