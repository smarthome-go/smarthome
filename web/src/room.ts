import type { ShallowDeviceResponse } from "./device"

export interface Room {
    data: {
        id: string
        name: string
        description: string
    }
    devices: ShallowDeviceResponse[]
    cameras: Camera[]
}

export interface Camera {
    id: string
    name: string
    url: string
    roomId: string
}

// The purpose of this function is to cache the device layout.
// This way, if the user visits the page after the initial load, there will be placeholders.
function storeDeviceLayout() {
    // TODO
}
