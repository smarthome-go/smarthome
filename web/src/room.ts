import type { DeviceResponse } from "./device"

export interface Room {
    data: {
        id: string
        name: string
        description: string
    }
    devices: DeviceResponse[]
    cameras: Camera[]
}

export interface Camera {
    id: string
    name: string
    url: string
    roomId: string
}
