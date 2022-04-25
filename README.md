# Smarthome
### Version: `0.0.25-beta-rc.1`

A completely self-built Smarthome-system written in Go

[![Go Build](https://github.com/MikMuellerDev/smarthome/actions/workflows/go.yml/badge.svg)](https://github.com/MikMuellerDev/smarthome/actions/workflows/go.yml)
[![](https://tokei.rs/b1/github/MikMuellerDev/smarthome?category=code)](https://github.com/MikMuellerDev/smarthome).

### The `setup.json`
Most of the configuration of the smarthome server can be achieved using the `setup.json` file.
This file is scanned and evaluated at startup.

```json
{
    "hardwareNodes": [
        {
            "name": "test raspberry pi",
            "url": "http://localhost:8070",
            "token": "smarthome"
        }
    ],
    "rooms": [
        {
            "id": "test_room",
            "name": "Test Room",
            "description": "This is a test room",
            "switches": [
                {
                    "id": "s1",
                    "name": "Lamp1"
                },
                {
                    "id": "s2",
                    "name": "Lamp2"
                }
            ],
            "cameras": [
                {
                    "name": "Test Camera",
                    "url": "https://mik-mueller.de/assets/foo.png"
                }
            ]
        }
    ]
}
```
