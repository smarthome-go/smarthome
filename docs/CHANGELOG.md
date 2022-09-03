## Changelog for v0.2.4

### Bugfixes
#### Web-GUI
- Fixed add-automation button misalignment on the automation page
- Added working sort-by-time to the automations and scheduler preview on the dashboard
- Added a time-until text to the automations on the dashboard
- Implemented missing dashboard GUI handling when an automation or schedule is currently executing
- Implemented missing Homescript argument data / display types in the argument prompt dialog
- Fixed issues regarding the HMS argument prompt

#### Server & Backend
- Modified the Homescript `print` function
- Previously, every argument to the function was seperated using a newline
- Now, the arguments are separated using whitespaces and the print function adds one newline at the end of all joined arguments
