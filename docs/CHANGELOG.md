## Changelog for v0.5.0

### Feature Additions

- Added a `factory reset` button to the system page

### Bugfixes

- Output of `print` will now also be streamed in the web-editor terminal
- Fixed a backend issue which caused errors when modifying automations whilst
  the automation system was disabled
- Once-disabled automations are now filtered out of the automations overview
- Dashboard tiles are now shown depending on the user's permissions
- Fixed the HMS argument prompt display when bool / yes-no is used
- HMS: Fixed the display of objects using the `debug` function

### Code Cleanup

- The Homescript web-editor now uses the official `codemirror-lang-homescript`
  package
