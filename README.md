# Smarthome
**Version**: `0.0.27-beta`

A completely self-built Smarthome-system written in Go.

[![Go](https://github.com/smarthome-go/smarthome/actions/workflows/go.yml/badge.svg)](https://github.com/smarthome-go/smarthome/actions/workflows/go.yml)
[![](https://tokei.rs/b1/github/smarthome-go/smarthome?category=code)](https://github.com/smarthome-go/smarthome).

## What is Smarthome?
Smarthome is a completely self-build home-automation system written in Go *(backend)* and Svelte *(frontend)*.
The system focuses on functionality and simplicity in order to guarantee a stable and reliable home-automation system which is actually helpful in automating common tasks.

### Concepts
- Completely self-hostable on your own infrastructure
- Simple setup: when version `1.0` is released, the entire configuration can be managed from the web interface
- Is able to operate without internet connection (except for the weather which relies on an API service)
- Privacy focused: Your data will stay on your system because Smarthome is not relying on cloud infrastructure
- An up-to-date docker-image is built and published to Docker Hub on every release 

## Hardware
As of April 27, 2022 the only way to make Smarthome interact with the real world is through the use of [node](https://github.com/smarthome-go/node), a Hardware interface which is required in order to interact with most generic 433mhz remote-sockets.
Naturally, the use of smarthome-hw requires physical hardware in order to communicate with remote sockets.

However, support for additional hardware, for example Zigbee devices is planned and would open additional possibilities for integration with other hardware.

## Getting Started
### Config.json
The config file can be used to setup database access, networking and debug mode.

**Docker:**
When running Smarthome via Docker, you might skip to the *setup.json* section 
```json
{
    "server": {
        "production": false,
        "port": 8082
    },
    "database": {
        "username": "smarthome",
        "password": "password",
        "database": "smarthome",
        "hostname": "localhost",
        "port": 3313
    }
}
```

### Setup.json
Basic parts of the configuration can be achieved using the `data/config/setup.json` file.
This file is scanned and evaluated at startup.

#### Example Configuration

```json
{
    "hardwareNodes": [
        {
            "name": "test raspberry pi",
            "url": "http://localhost:80",
            "token": "secret_token"
        }
    ],
    "rooms": [
        {
            "data": {
                "id": "test",
                "name": "Test Room",
                "description": "This is a test"
            },
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
                    "id": "test_camera",
                    "name": "Test Camera",
                    "url": "https://mik-mueller.de"
                }
            ]
        }
    ]
}
```
