## Changelog for v0.2.0

### Bugfixes
- Decrease label size of create-new HMS button in HMS manager page (*avoid text-wrapping on mobile*)
- Decreased label text ambiguity of the same create-new HMS button
- Fixed overflowing header when using too many HMS-workspaces
- Updated some `NPM` dependencies (*for the web-UI*)

### Additions
- Added the `/api/power/usage/all` API endpoint which returns all power-usage data since records started
- This endpoint is protected with normal API-authentication because it will cause some load on the server when requested 
