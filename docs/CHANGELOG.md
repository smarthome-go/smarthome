## Changelog for v0.0.58

### Dashboard
- Added *Quick Actions* as a dashboard component
- Improved the dashboard's reactivity

#### Power Usage Monitor
- Added the unit `Watts` to the tooltip label
- Removed line tension which leads to *more sharp edges*
- Optimized graph scaling on non-widescreen devices
- Decreased measurement data point radii
- Improved the graph's fill gradient to be less dominant

### Power Usage Data Collection
- Added a safety check which prevents errors from happening during the calculation of the power usage
- This bug could occur if you only had switches which each 'used' `0` Watts
- Added a scheduler which performs a periodic power usage snapshot (*every hour*)
- This should help make the graph more accurate and should avoid the generation of misleading slopes
- Added data filtering and automated improvement of the power usage data points
- Redundant measurements are now automatically removed from the records
- This will improve the chart's visuals and will make it more organized (*Easier to spot real changes*)

### Server Improvements
- Fixed non-running HMS URL cache flushing scheduler
- Fixed a bug in the Homescript runtime environment
    - When using an `exec` call, termination of the Homescript would not work properly
    - Only the *inner* Homescript, the `exec` target would get terminated
    - This issue is now resolved, however `exec` will return an error if terminated

