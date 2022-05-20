import { get, writable, Writable } from 'svelte/store'
import { createSnackbar } from '../../global'
import App from './App.svelte'


export interface User {
  username: string
  forename: string
  surname: string
  primaryColorDark: string
  primaryColorLight: string
  schedulerEnabled: boolean
  darkTheme: boolean
}

export interface Permission {
  permission: string
  name: string
  description: string
}

export interface Switch {
  id: string
  name: string
  roomId: string
  powerOn: boolean
  watts: number
}

export interface UserData {
  user: User
  permissions: string[]
  switchPermissions: string[]
}

export interface Camera {
  id: string
  name: string
  url: string
  roomId: string
}


export const loading: Writable<boolean> = writable(false)
export const users: Writable<UserData[]> = writable([])
export const allPermissions: Writable<Permission[]> = writable([])
export const allSwitches: Writable<Switch[]> = writable([])
export const allCameras: Writable<Camera[]> = writable([])
export const allSwitchesFetched: Writable<boolean> = writable(false)
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

export async function fetchAllSwitches() {
  try {
    const res = await (await fetch('/api/switch/list/all')).json()
    if (res.success !== undefined && !res.success) throw Error(res.error)
    allSwitches.set(res)
  } catch (err) {
    get(createSnackbar)(`Could not load system switches: ${err}`)
  }
  allSwitchesFetched.set(true)
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