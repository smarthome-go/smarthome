import type { ConfigAction } from '@smui/snackbar/kitchen'
import { get, Writable, writable } from 'svelte/store'

export const createSnackbar: Writable<(message: string, actions?: ConfigAction[]) => void> = writable(() => { })

export interface Notification {
    id: number
    priority: number
    name: string
    description: string
    date: string
}

export interface Data {
    userData: UserData
    notificationCount: number
    notifications: Notification[]
    loaded: boolean
}

export interface UserData {
    user: {
        username: string
        forename: string
        surname: string
        primaryColorDark: string
        primaryColorLight: string
        darkTheme: boolean
    }
    permissions: string[]
}

// Color caching
let cachedColorDark = localStorage.getItem("smarthome_primary_color_dark")
let cachedColorLight = localStorage.getItem("smarthome_primary_color_light")

if (cachedColorDark !== null && cachedColorLight !== null) {
    document.documentElement.style.setProperty('--clr-primary-dark', cachedColorDark)
    document.documentElement.style.setProperty('--clr-primary-light', cachedColorLight)
}

export const data: Writable<Data> = writable({
    userData: {
        user: {
            forename: '',
            primaryColorDark: cachedColorDark,
            primaryColorLight: cachedColorLight,
            surname: '',
            username: '',
            darkTheme: true,
        },
        permissions: [],
    },
    notificationCount: 0,
    notifications: [],
    loaded: false,
})

let isFetching = false
let hasFetched = false // Indicates that the user data has been fetched, used for primary color caching

export async function fetchData() {
    if (hasFetched) return
    if (isFetching) {
        while (isFetching) await sleep(5)
        return
    }
    isFetching = true
    const temp = get(data)
    temp.userData = await fetchUserData()
    temp.notificationCount = await fetchNotificationCount()
    temp.loaded = true
    data.set(temp)
    isFetching = false
    hasFetched = true

    // Update cached primary colors
    localStorage.setItem("smarthome_primary_color_dark", get(data).userData.user.primaryColorDark)
    localStorage.setItem("smarthome_primary_color_light", get(data).userData.user.primaryColorLight)
}

export async function fetchUserData(): Promise<UserData> {
    try {
        const res = await (await fetch('/api/user/data')).json()
        if (res.success !== undefined && !res.success) throw Error(res.error)
        return res
    } catch (err) {
        get(createSnackbar)(`Could not fetch user data: ${err}`)
    }
}

export async function fetchNotificationCount(): Promise<number> {
    try {
        const res = await (await fetch('/api/user/notification/count')).json()
        if (res.success !== undefined && !res.success) throw Error(res.error)
        return res.count
    } catch (err) {
        get(createSnackbar)(`Could not fetch notification count: ${err}`)
    }
}

export const sleep = (ms: number) => new Promise((res) => setTimeout(res, ms))
