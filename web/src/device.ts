import type { ConfigSpecWrapper } from "./driver"
import type { homescriptError } from 'src/homescript';
import { get } from "svelte/store";
import { createSnackbar } from "./global";

export type DeviceType = 'INPUT' |'OUTPUT'

export interface HydratedDeviceResponse {
    shallow: ShallowDeviceResponse;
    extractions: DeviceExtractions;
}

export interface DeviceExtractions {
    hmsErrors: homescriptError[];
    config: ConfigSpecWrapper,
    powerInformation: DevicePowerInformation,
    dimmables: DeviceDimmable[],
    sensors: DeviceSensor[],
}

export interface ShallowDeviceResponse {
    type: DeviceType
    id: string
    name: string
    roomId: string
    vendorId: string,
    modelId: string,
    singletonJson: {},
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

export interface DeviceSensor {
    label: string,
    value: {},
    hmsType: string,
    unit: string,
}

export interface CreateDeviceRequest {
    type: DeviceType
    id: string
    name: string
    roomId: string
    driverVendorId: string,
    driverModelId: string,
}

export interface ModifyDeviceRequest {
    id: string,
    name: string,
}

export async function fetchAllShallowDevices(): Promise<ShallowDeviceResponse[]> {
    try {
        let res = await fetch("/api/devices/list/all")

        let resJson = await res.json()
        if (resJson.success === false) {
            throw(resJson.error)
        }

        return resJson as ShallowDeviceResponse[]
    } catch (err) {
        get(createSnackbar)(`Could not load devices: ${err}`)
    }
}
