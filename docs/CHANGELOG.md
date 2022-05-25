## Changelog for v0.0.34-fix.1

### Homescript
- The `http` function now also requires the `Content-Type` agument
- Code using the new features: `http('http://localhost', 'POST', 'application/json', '{"id": 1}')`

### Development
- Added additional integration test for the camera and camera-permission feature of the database
