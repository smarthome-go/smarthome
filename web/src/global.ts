import type { SnackbarComponentDev } from '@smui/snackbar'
import { Writable, writable, get } from 'svelte/store'

export const infoBar: Writable<{
    message: string,
    bar: SnackbarComponentDev,
}> = writable({
    message: '',
    bar: undefined,
})

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
}

export interface UserData {
    username: string
    forename: string
    surname: string
    primaryColor: string
    darkTheme: boolean
}

export const data: Writable<Data> = writable({
    userData: {
        forename: '',
        primaryColor: '',
        surname: '',
        username: '',
        darkTheme: true,
    },
    notificationCount: 0,
    notifications: []
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
    data.set(temp)
    console.log('Fetched data:', temp)
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
