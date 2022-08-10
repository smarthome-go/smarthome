## Changelog for v0.0.55

### Token Authentication
- The `API` endpoint at `/api/login/token` now returns information about the token that was used to authenticate (*only if the token was valid*)
- This release does not add a lot of features but is required for maintaining compatibility with [`SDK`](https://github.com/smarthome-go/sdk) version `0.19.0`.

#### Example
##### Request
`POST` to `/api/login/token` with following body
```json
{
    "token": "650feaafc1487d18bd8c5a805363be96"
}
```

##### Returns
`200 OK`
```json
{
    "username": "test",
    "tokenLabel": "myapp"
}
```
