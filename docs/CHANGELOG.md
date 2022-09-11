## Changelog for v0.2.7

### Bugfixes
#### Server & Backend
- Fixed a bug which affected the scheduler
- Even if a schedule was executed, it was not removed from the database, thus creating a misleading display of the system's crrent state
