import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface Room {
    data: {
        id: string
        name: string
        description: string
    }
    switches: SwitchResponse[]
    cameras: Camera[]
}
export interface SwitchResponse {
    id: string
    name: string
    powerOn: boolean
    watts: number
}

export interface Camera {
    id: string
    name: string
    url: string
    roomId: string
}

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
