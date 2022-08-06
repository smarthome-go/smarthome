## Changelog for v0.0.51

### Scheduler GUI
- Improved the scheduler GUI significantly
- Fixed several bugs (including the real-time execution count-down)
- Added important information displays to the individual schedule panels
- ⇒ *Note*: The GUI is now in a working and pretty state

### Automations GUI
- Improved the automations GUI
- Modified the width of each individual automation item to align well on `1080p`
- Made the automation's title appear bold

### Homescript GUI
- Fixed broken execution result popup
- Fixed broken argument prompts
- ⇒ *Note*: These bugfixes also affect the *automations GUI* and the *Homescript manager's GUI*

### Homescript Editor
- Added argument prompts when running or linting using the editor's *"terminal"*

### Homescript Backend
#### Arguments During Linting
- Attempt to use provided arguments when using lint
- If an argument is specified, it is used during linting
- If it is omitted, a default (*empty string*) is used in order to prevent errors

#### SigTerm Handling
- If `exec` is used, the executor's `sigTerm` is forwarded to the *exec target*
- This means that a script termination will work as expected when using `exec` in your code

### Bugfixes / Code Quality
- Fixed several typos in the HMS deletion dialog
- Improved logging in the backend *automation* module
