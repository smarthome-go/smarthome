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

export interface UserData {
  user: User
  permissions: string[]
}

export const users: Writable<UserData[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export const allPermissions: Writable<Permission[]> = writable([])

export async function fetchAllPermissions() {
  try {
    const res = await (await fetch('/api/permissions/list')).json()
    if (res.success !== undefined && !res.success) throw Error(res.error)
    allPermissions.set(res)
  }catch(err) {
    get(createSnackbar)(`Could not fetch system permissions: ${err}`)
  }
}

export default new App({
  target: document.body,
})