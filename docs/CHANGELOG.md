## Changelog for v0.0.57

### Dashboard
- Started work on the dashboard
- Added the first component: a power-usage graph

### System Configuration Page
- Fixed edit-button label in the hardware-node edit dialog

### Homescript
- Upgraded [Homescript](https://github.com/smarthome-go/homescript/releases/tag/v0.16.0) to `v0.17.0`
- Then to [`v0.16.0`](https://github.com/smarthome-go/homescript/releases/tag/v0.16.0)
- Added the `currentWeek` variable the `even` function and the `ping` function
##### Getting the Current Week
- Returns the ISO 8601 week number of the current year
- Ranges from 1 to 53

```python
print(currentWeek)
```

##### Checking if a Number is Even
- Takes an integer as its input (*floats will be implicitly converted to int*)
- Checks if the parameter is even and returns an according boolean value

```python
print(even(1))
```

##### Performing an ICMP Ping
- Requires a hostname and a timeout (*in seconds*)
- Returns a boolean indicating whether the requested host is online or offline

```python
print(ping('localhost', 0.5))
```

### Server Backend
- Added all necessary functions to support the collection of power usage data
- On every switch's power change, a snapshot of the current power usage is taken and saved into the database
