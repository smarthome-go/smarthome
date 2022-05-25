## Changelog for v0.0.35

### Homescript
- Modified the Homescript implementation to support a `dry-run` setting which disables:
    - Data manipulation
    - Extended waiting opportunities for data, such as network requests
- During lint, Homescript still checks arguments and even function-specific values, such as the existence of a switch or the possession of a permission
- Added the `/api/homescript/lint/live` API endpoint which is used to lint Homescript code
