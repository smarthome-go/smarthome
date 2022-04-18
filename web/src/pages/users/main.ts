import { writable, Writable } from 'svelte/store'
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

export const allPermissions: Writable<Permission[]> = writable([])

export default new App({
  target: document.body,
})