# Quickstart
## Pre-Installation
### Config.json
The config file can be used to set up database access, networking and debug mode.

**Docker:**
When running Smarthome via Docker, you might want to skip to the *setup.json* section.
Because all parameters of the `config.json` are configurable via environment variables, a cleaner way to configure Smarthome with Docker would be through the use of `docker-compose.yml`.
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
Basic parts of the configuration can be achieved using the `setup.json` file, which is located at `data/config/setup.json`.
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
## Installation
### Getting Started Via Docker
The quickest way to get started with Smarthome is by using Docker.
#### Unfortunately, Docker is currently only supported for AMD64 devices. However, support for other architectures, such as ARM is planned.
1. Make sure an up-to-date version of both Docker and docker-compose is installed on the target machine.
2. Copy the contents of [`docker-compose.yml`](../docker-compose.yml) or download it.
3. When aiming for a system which can be reproduced easily, you may want to have a look at the [`setup.json`](#Setup.json) file first.
3. Run `docker-compose up [-d]` in order to start the service.

