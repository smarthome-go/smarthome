import { createSnackbar } from "../../global";
import { get } from "svelte/store";

export interface DriverData {
    vendorId: string,
    modelId: string,
    name: string,
    version: string,
}

export interface CreateDriver {
    data: DriverData,
    code: string,
}

export interface FetchedDriver {
    vendorId: string,
    modelId	: string,
    name: string,
    version: string,
    homescriptId: string,
}


export async function fetchDrivers() {
    try {
        const res = await (
            await fetch("/api/system/hardware/drivers/list")
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
            await fetch("/api/system/hardware/drivers/add", {
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
