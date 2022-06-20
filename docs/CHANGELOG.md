## Changelog for v0.0.45

### Homescript GUI
- Revised the management layout so that the action buttons (delete / pick icon) seem less dominant
- Improved Homescript run result dialog display for mobile devices
- Added *Edit code* button to the actions which opens the HMS IDE

#### Homescript IDE
- Added an initial version of a web-based Homescript editor / IDE
- Code can now be edited via the web interface
- Support for immature syntax highlighting (*currently using an incorrect grammar definition*)
- Improved `HmsEditor.svelte` component which is used by the IDE
- Split layout: the layout is separated into the text editor and the terminal / output area
  - The layout's proportions are `25% | 75%` (can be swapped to `75% | 25%`)
- Includes a small script selection menu at the top: allows easy switching between scripts
- The active script is set using `URL params`, for example `http://smarthome.box/homescript/editor?id=test`
