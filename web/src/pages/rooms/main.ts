import { type Writable, writable } from 'svelte/store'
import App from './App.svelte'
import type { ConfigSpecWrapper } from '../../driver'
import type { homescriptError } from 'src/homescript'


export const requests: Writable<number> = writable(0)

// Specifies whether the cameras will reload every 10 seconds
export const periodicCamReloadEnabled: Writable<boolean> = writable(
    localStorage.getItem('smarthome_periodic_cam_reload_enabled') === 'true',
)
// Specifies whether
export const powerCamReloadEnabled: Writable<boolean> = writable(
    localStorage.getItem('smarthome_power_cam_reload_enabled') === 'true',
)

export default new App({
    target: document.body,
})
