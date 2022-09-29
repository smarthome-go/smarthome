import { writable, type Writable } from 'svelte/store'
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

// Specifies whether the cameras will reload every 10 seconds
export const periodicCamReloadEnabled: Writable<boolean> = writable(localStorage.getItem("smarthome_periodic_cam_reload_enabled") === "true")
// Specifies whether
export const powerCamReloadEnabled: Writable<boolean> = writable(localStorage.getItem("smarthome_power_cam_reload_enabled") === "true")

export default new App({
    target: document.body,
})
