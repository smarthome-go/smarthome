import { createSnackbar } from "../../global";
import { get } from "svelte/store";

export interface CreateDriver {
    data: DriverData,
    code: string,
}

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
}

export interface DriverInfoCollection {
    driver: ConfigSpecWrapper;
    device: ConfigSpecWrapper;
}

export interface ConfigSpecWrapper {
    config: ConfigSpec
    // NOTE: hms type from backend ignored here, would add unnessecary bloat and is not needed here.
}

export type ConfigSpec = ConfigSpecAtom | ConfigSpecInner | ConfigSpecStruct | null;

export type ConfigSpecType = 'INT' | 'FLOAT' | 'BOOL' | 'STRING' | 'LIST' | 'STRUCT' | 'OPTION'

export interface ConfigSpecAtom {
    type:   ConfigSpecType;
}

export interface ConfigSpecInner {
    type:   ConfigSpecType;
    inner: ConfigSpec;
}

export interface ConfigSpecStruct {
    type: ConfigSpecType;
    fields: ConfigSpecStructField[];
}

export interface ConfigSpecStructField {
    name: ConfigSpecType;
    type: ConfigSpec;
}

export interface ValidationError {
    level:   number;
    message: string;
    notes:   string[];
    span:    Span;
}

export interface Span {
    start:    Location;
    end:      Location;
    filename: string;
}

export interface Location {
    line:   number;
    column: number;
    index:  number;
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
export async function createDriver(
    data: CreateDriver
) {
    try {
        const res = await (
            await fetch("/api/system/hardware/driver/add", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    data,
                    code: "// Enter device Driver code here\n"
                }),
            })
        ).json();
        if (res.success !== undefined && !res.success)
            throw Error(res.error);
    } catch (err) {
        get(createSnackbar)(`Failed to create hardware driver node: ${err}`);
    }
}
