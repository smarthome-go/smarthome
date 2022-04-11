import type { ConfigAction } from '@smui/snackbar/kitchen'
import { get, Writable, writable } from 'svelte/store'

export const createSnackbar: Writable<(message: string, actions?: ConfigAction[]) => void> = writable(() => {})

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
    username: string
    forename: string
    surname: string
    primaryColorDark: string
    primaryColorLight: string
    darkTheme: boolean
}

export const data: Writable<Data> = writable({
    userData: {
        forename: '',
        primaryColorDark: '',
        primaryColorLight: '',
        surname: '',
        username: '',
        darkTheme: true,
    },
    notificationCount: 0,
    notifications: [],
    loaded: false,
})

let isFetching = false
let hasFetched = false

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
}

export async function fetchUserData(): Promise<UserData> {
    return await (await fetch('/api/user/data')).json()
}

export async function fetchNotificationCount(): Promise<number> {
    return (await (await fetch('/api/user/notification/count')).json()).count
}

export const sleep = (ms: number) => new Promise((res) => setTimeout(res, ms))
