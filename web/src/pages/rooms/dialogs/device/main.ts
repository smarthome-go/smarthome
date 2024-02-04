import { get, writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import { createSnackbar } from '../../../../global'
import { fetchDrivers as fetchDriversInternal, type FetchedDriver } from '../../../system/driver'

export const loading: Writable<boolean> = writable(true)
export const hardwareNodes: Writable<HardwareNode[]> = writable([])
export const hardwareNodesLoaded: Writable<boolean> = writable(false)

export interface HardwareNode {
    url: string
    name: string
    token: string
    enabled: boolean
    online: boolean
}

export async function fetchHardwareNodes() {
    loading.set(true)
    try {
        const res = await (await fetch(`/api/system/hardware/node/list/nopriv`)).json()
        if (res.success !== undefined && !res.success) throw Error(res.error)
        hardwareNodes.set([...res, null])
        hardwareNodesLoaded.set(true)
    } catch (err) {
        get(createSnackbar)(`Failed to load hardware nodes: ${err}`)
    }
    loading.set(false)
}

export const drivers: Writable<FetchedDriver[]> = writable([])
export const driversLoaded: Writable<boolean> = writable(false)

export interface HardwareNode {
    url: string
    name: string
    token: string
    enabled: boolean
    online: boolean
}

export async function fetchDrivers() {
    loading.set(true)
    let driversTemp =  await fetchDriversInternal()
    drivers.set(driversTemp)
    driversLoaded.set(true)
    loading.set(false)
}
