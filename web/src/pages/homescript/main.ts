import { writable, Writable } from 'svelte/store'
import App from './App.svelte'
import type {homescript} from "../../homescript"


// States that homescripts have been loaded
// used when trying to access the data of the automation's homescript
export const hmsLoaded: Writable<boolean> = writable(false)
export const homescripts: Writable<homescript[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
