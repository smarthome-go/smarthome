import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface homescript {
    owner: string
    data: homescriptData 
}

export interface homescriptData {
    id: string
    name: string
    description: string
    mdIcon: string
    code: string
    quickActionsEnabled: boolean
    schedulerEnabled: boolean
}

// States that homescripts have been loaded
// used when trying to access the data of the automation's homescript
export const hmsLoaded: Writable<boolean> = writable(false)
export const homescripts: Writable<homescript[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
