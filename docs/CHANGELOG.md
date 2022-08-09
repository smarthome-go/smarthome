## Changelog for v0.0.53

### Profile Page GUI
- A user can now change their password using the button under the *Danger Zone*
- This will sign out the user immediately and will therefore force the user to remember their password
- Significantly improved the layout on different viewports, for example `widescreen` and `mobile`
- Minimized layout shifts during page load
- Added `authentication tokens` and their management to the profile page
- Improved avatar image reset button location using a dynamic popup
- The profile page is now officially usable and almost completed

### Authentication Tokens
- As described above, authentication tokens have been added to Smarthome
- In the future, the `SDK` will add support for tokens in order to increase the security of the Smarthome infrastructure

### Homescript
- Removed status code checking during linting of `http` functions

### User Permissions
- Renamed the permission `Logs` to `Event Logs`

### Code Quality / Bugfixes
- Removed unneeded `uui` NPM dependency
- 

