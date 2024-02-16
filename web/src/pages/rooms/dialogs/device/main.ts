import { writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import { fetchDrivers as fetchDriversInternal, type FetchedDriver } from '../../../system/driver'

export const loading: Writable<boolean> = writable(true)
export const drivers: Writable<FetchedDriver[]> = writable([])
export const driversLoaded: Writable<boolean> = writable(false)

export async function fetchDrivers() {
    loading.set(true)
    let driversTemp =  await fetchDriversInternal()
    drivers.set(driversTemp)
    driversLoaded.set(true)
    loading.set(false)
}
