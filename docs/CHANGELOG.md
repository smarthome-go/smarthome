## Changelog for v0.0.49

### Scheduler
- The scheduler web-page is now accessible under `/scheduler`
- Added a switch *wizard* to the scheduler's target selector
- Refactored entire front- and backend to use *structured* data instead of just an HMS code field

### Performance
- Used *Vite's* manual chunking for better performance

### Bugfixes
- Dangling automations and schedules will now be removed on their next run
- Added a check if *schedule-selection* is enabled to automation and scheduler API
- Fixed many typos
