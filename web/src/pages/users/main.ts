import { get, writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import { createSnackbar, type ShallowUserData } from '../../global'
import { fetchAllShallowDevices, type ShallowDeviceResponse } from '../../device'
import App from './App.svelte'
import type { Camera } from 'src/room'

export interface Permission {
    permission: string
    name: string
    description: string
}

export interface PermissionUserData {
    user: ShallowUserData
    permissions: string[]
    devicePermissions: string[]
}

export const loading: Writable<boolean> = writable(false)
export const users: Writable<PermissionUserData[]> = writable([])
export const allPermissions: Writable<Permission[]> = writable([])
export const allDevices: Writable<ShallowDeviceResponse[]> = writable([])
export const allCameras: Writable<Camera[]> = writable([])
export const allDevicesFetched: Writable<boolean> = writable(false)
export const allCamerasFetched: Writable<boolean> = writable(false)

export async function fetchAllPermissions() {
    try {
        const res = await (await fetch('/api/permissions/list/all')).json()
        if (res.success !== undefined && !res.success) throw Error(res.error)
        allPermissions.set(res)
    } catch (err) {
        get(createSnackbar)(`Could not load system permissions: ${err}`)
    }
}

export async function fetchAllDevices() {
    try {
        // const res = await (await fetch('/api/switch/list/all')).json()
        // if (res.success !== undefined && !res.success) throw Error(res.error)
        const res = await fetchAllShallowDevices()
        allDevices.set(res)
    } catch (err) {
        get(createSnackbar)(`Could not load system devices: ${err}`)
    }
    allDevicesFetched.set(true)
}

export async function fetchAllCameras() {
    try {
        const res = await (await fetch('/api/camera/list/redacted')).json()
        if (res.success !== undefined && !res.success) throw Error(res.error)
        allCameras.set(res)
    } catch (err) {
        get(createSnackbar)(`Could not load system cameras: ${err}`)
    }
    allCamerasFetched.set(true)
}

export default new App({
    target: document.body,
})
