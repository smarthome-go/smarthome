import { createSnackbar } from "../../global";
import { get } from "svelte/store";
import type { ConfigSpecWrapper, ValidationError } from "../../driver";

export type CreateDriverReq = DriverData & { homescriptCode: null | string }

export interface FetchedDriver {
    driver:           DriverData;
    info:             DriverInfoCollection;
    configuration:    {},
    isValid:          boolean;
    validationErrors: ValidationError[];
}

export interface DriverData {
    vendorId:       string;
    modelId:        string;
    name:           string;
    version:        string;
    homescriptCode: string;
    dirty: boolean;
}

export interface DriverInfoCollection {
    driver: ConfigSpecWrapper;
    device: ConfigSpecWrapper;
}

export async function fetchDrivers(): Promise<FetchedDriver[]> {
    try {
        const res = await (
            await fetch("/api/system/hardware/driver/list")
        ).json();
        if (res.success !== undefined && !res.success)
            throw Error(res.error);
        return res
    } catch (err) {
        get(createSnackbar)(`Failed to load hardware drivers: ${err}`);
    }
}

// Creates a new hardware node
export async function createDriver(data: CreateDriverReq) {
    try {
        const res = await (
            await fetch("/api/system/hardware/driver/add", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            })
        ).json();
        if (res.success !== undefined && !res.success)
            throw Error(res.error);
    } catch (err) {
        get(createSnackbar)(`Failed to create hardware driver node: ${err}`);
    }
}
