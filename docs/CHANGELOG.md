## Changelog for v0.0.52

### User Profile GUI
- Added a first version of the user profile GUI
#### Features
- User data visualization
- User data modification
- Avatar display
- Avatar deletion and modification
- User account deletion

### Homescript
- Bumped the Homescript runtime to `v0.15.2`
- Therefore fixed a possible server panic when using the builtin `http` function

#### Networking
- Homescript can now use the `sigTerm` in order to cancel the builtin `http` and `get` functions during a request
- Significantly improved `lint` times when using *networking* inside Homescript
- Implemented a caching-system which caches correctly linted URLs inside a database table
- Each cache entry expires after `12 hours`, thus keeping each URL up-to-date and reducing database table size

### Backend Code Quality
- Refactored a large portion of the backend user avatar logic
- Refactored user avatar database functions to avoid unfitting code
- Upgraded all used `Go` modules

### Developer Notes
- Removed useless test in `cpre/config/config_test.go`
- Bumped dependency `vite` to `^3.0.0` and therefore fixed broken `npm i`
- Removed annoying test output from `core/config/export_test.go`

### Documentation
- Added hint for session keys to `Docker.md`

### GUI fixes
- Eliminated large content shift in NavBar pages
- Added missing data attributes to the global data store
- Made permission names and descriptions more concise and readable
- Fixed scrollable room page on non-widescreen layouts
