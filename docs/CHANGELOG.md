## Changelog for v0.7.0

- Added HMS widgets
  - Homescripts can now be used as dashboard widgets
  - The setting `Is Widget` of the desired Homescript is to be activated
  - Everything outputted via `print` or `println` is rendered as bare HTML
  - Javascript is also supported
- Fixed various Homescript-related bugs
- The automation and scheduling system can now be controlled via Homescript

```lua
# Example
scheduler.new({
  name: "Title",
  hour: 4,
  minute: 21,
  code: "switch('switch_id', on);"
});
  ```
