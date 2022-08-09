## Changelog for v0.0.54

### Token Authentication
- Services can now use the `API` endpoint at `/api/login/token` to authenticate themselves using a token
- This release does not add a lot of features but is required for maintaining compatibility with the [`SDK`](https://github.com/smarthome-go/sdk).

#### Example Request
`POST` to `/api/login/token` with following body
```json
{
    "token": "650feaafc1487d18bd8c5a805363be96"
}
```
