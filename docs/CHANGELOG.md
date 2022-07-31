## Changelog for v0.0.50

### Scheduler
- Improved the Homescript selector UI, this also affects the *automations GUI* positively
- Made several smaller tweaks to the scheduler UI
- Added an indicator which displays the time until the schedule's next execution
- Added a meaningful helper text which also shows the next execution time
- Added automatic deletion of running schedules
- When the frontend notices an executing schedule, it is removed from the UI

### Automations
- Improved the Homescript selector UI, this also affects the *scheduler GUI* positively
- Added a meaningful helper text to the time picker which displays the time until the automation's next execution

### Backend
- Added the `SMARTHOME_SESSION_KEY` environment variable
- For more information about this parameter, read [this documentation](./Docker.md).
- Added HMS job count to debug info


### Bugfixes / Code Quality
- All known typos in the source code have been fixed
- Added automatic typo checks via *Github actions*
- Improved the Docker-specific documentation significantly
