import { type Writable, writable } from 'svelte/store'
import App from './App.svelte'
import type { ConfigSpecWrapper } from '../../driver'
import type { homescriptError } from 'src/homescript'

export interface Room {
    data: {
        id: string
        name: string
        description: string
    }
    devices: DeviceResponse[]
    cameras: Camera[]
}

export type DeviceType = 'INPUT' |'OUTPUT'

export interface DeviceResponse {
    type: DeviceType
    id: string
    name: string
    roomId: string
    vendorId: string,
    modelId: string,
    singletonJson: {},
    hmsErrors: homescriptError[];
    config: ConfigSpecWrapper,
    powerInformation: DevicePowerInformation,
    dimmables: DeviceDimmable[],
}

export interface DevicePowerInformation {
    state: boolean,
    powerDrawWatts: number,
}

export interface DeviceDimmableRange {
    lower: number,
    // Upper is always exclusive.
    upper: number,
}

export interface DeviceDimmable {
    value: number,
    label: string,
    range: Range
}

export interface CreateDeviceRequest {
    type: DeviceType
    id: string
    name: string
    roomId: string
    driverVendorId: string,
    driverModelId: string,
}

export interface Camera {
    id: string
    name: string
    url: string
    roomId: string
}

export const loading: Writable<boolean> = writable(false)

// Specifies whether the cameras will reload every 10 seconds
export const periodicCamReloadEnabled: Writable<boolean> = writable(
    localStorage.getItem('smarthome_periodic_cam_reload_enabled') === 'true',
)
// Specifies whether
export const powerCamReloadEnabled: Writable<boolean> = writable(
    localStorage.getItem('smarthome_power_cam_reload_enabled') === 'true',
)

export default new App({
    target: document.body,
})
