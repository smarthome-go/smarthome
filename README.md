# Smarthome
**Version**: `0.0.55`

A completely self-built Smarthome-system written in Go.

[![Go Build](https://github.com/smarthome-go/smarthome/actions/workflows/go.yml/badge.svg)](https://github.com/smarthome-go/smarthome/actions/workflows/go.yml)
[![Check for Typos](https://github.com/smarthome-go/smarthome/actions/workflows/typos.yml/badge.svg)](https://github.com/smarthome-go/smarthome/actions/workflows/typos.yml)

## What is Smarthome?
Smarthome is a completely self-build home-automation system written in Go *(backend)* and Svelte *(frontend)*.
The system focuses on functionality and simplicity in order to guarantee a stable and reliable home-automation system which is actually helpful in automating common tasks.

### Concepts
- Completely self-hostable on your own infrastructure
- Simple setup: when server version `1.0` is released, the entire configuration will be manageable from the web interface
- Is able to operate without internet connection (except for the weather which relies on an API service)
- Privacy focused: Your data will stay on your system because Smarthome is not relying on cloud infrastructure
- An up-to-date docker-image is built and published to Docker-Hub on every release → *Official Docker support*
- Most of the configuration can be `exported` and `imported` via `JSON` → Reproducible configuration after web-setup

## Hardware
As of April 27, 2022 the only way to make Smarthome interact with the real world is through the use of [node](https://github.com/smarthome-go/node), a Hardware interface which is required in order to interact with most generic 433mhz remote-sockets.
Naturally, the use of node requires physical hardware in order to communicate with remote sockets.

However, support for additional hardware, for example Zigbee devices is planned and would open additional possibilities, for example integration with other hardware.

## Getting Started
A guide for getting started can be found [here](./docs/Quickstart.md).
