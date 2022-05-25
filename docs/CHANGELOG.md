## Changelog for v0.0.34

### Homescript
- Added Http methods for making network requests via Homescript
- Introduced the `hmsNetwork` permission which is required in order to make network requests from Homescript
- The `get` method requires an arbitrary URL and returns the response as a string
- An example for the print function is: `print(get('http://localhost:8082'))`
- The `http` method requires a URL, a request-method, and a request-body and returns the response as a string
- An example fot the http function is: `print(http('http://localhost:8082', 'POST', '{"key": "value"}'))`

### Tweaks
- Tweaked label of *create-automation* button in automation GUI
- Altered a debug message in the configuration backend
- Indroduced dynamic orientations when opening the web-ui as a PWA

### Docker-Container
- Created a symlink from `/usr/bin/homescript` to `/usr/bin/shome` 
- Allows the user to use the `shome` command as an alias for the `homescript` command
