import type { ConfigAction } from '@smui/snackbar/kitchen'
import { get, Writable, writable } from 'svelte/store'

// eslint-disable-next-line @typescript-eslint/no-empty-function 
export const createSnackbar: Writable<(message: string, actions?: ConfigAction[]) => void> = writable(() => { })

export interface GenericResponse {
    success: boolean,
    message: string,
    error: string,
    time: number,
}

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

// Given an arbitrary input color, the function decides whether text on the color should be white or black
export function contrast(color: string): 'black' | 'white' {
    const r = parseInt(color.slice(1, 3), 16)
    const g = parseInt(color.slice(3, 5), 16)
    const b = parseInt(color.slice(5, 7), 16)
    const a = [ r, g, b ].map(v => {
        v /= 255
        return v <= 0.03928
            ? v / 12.92
            : Math.pow((v + 0.055) / 1.055, 2.4)
    })
    const luminance = a[ 0 ] * 0.2126 + a[ 1 ] * 0.7152 + a[ 2 ] * 0.0722
    const [ darker, brighter ] = [ 1.05, luminance + 0.05 ].sort()
    return brighter / darker <= 4.5 ? 'black' : 'white'
}

// Color caching
const cachedColorDark: string = localStorage.getItem("smarthome_primary_color_dark")
const cachedColorLight: string = localStorage.getItem("smarthome_primary_color_light")
let cachedTheme = true
const cachedThemeTemp = localStorage.getItem("smarthome_dark_theme_enabled") === 'true'

if (cachedColorDark !== null && cachedColorLight !== null && cachedTheme !== null) {
    document.documentElement.style.setProperty('--clr-primary-dark', cachedColorDark)
    document.documentElement.style.setProperty('--clr-primary-light', cachedColorLight)

    document.documentElement.style.setProperty(
        '--clr-on-primary-dark',
        contrast(cachedColorDark) === 'black'
            ? '#121212'
            : '#ffffff'
    )
    document.documentElement.style.setProperty(
        '--clr-on-primary-light',
        contrast(cachedColorLight) === 'black'
            ? '#121212'
            : '#ffffff'
    )

    document.documentElement.classList.toggle('light-theme', cachedTheme)
    cachedTheme = cachedThemeTemp
}

export const data: Writable<Data> = writable({
    userData: {
        user: {
            forename: '',
            primaryColorDark: cachedColorDark,
            primaryColorLight: cachedColorLight,
            surname: '',
            username: '',
            darkTheme: cachedTheme,
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
    localStorage.setItem("smarthome_dark_theme_enabled", get(data).userData.user.darkTheme ? 'true' : 'false')
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

export async function hasPermission(permission: string): Promise<boolean> {
    while (!hasFetched) await sleep(5)
    const permissions = get(data).userData.permissions
    return (permissions.includes(permission) || permissions.includes('*'))
}
