import type { ConfigSpecWrapper } from "./driver"
import type { homescriptError } from 'src/homescript';

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
    sensors: DeviceSensor[],
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
