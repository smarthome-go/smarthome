import { writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import type { homescriptJob, homescriptWithArgs } from '../../homescript'
import App from './App.svelte'
import '@fontsource/jetbrains-mono'

// States that homescripts have been loaded
// used when trying to access the data of the automation's homescript
export const hmsLoaded: Writable<boolean> = writable(false)
export const homescripts: Writable<homescriptWithArgs[]> = writable([])
export const jobs: Writable<homescriptJob[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export const RESERVED_HOMESCRIPTS = [
    'sys',
]

export default new App({
    target: document.body,
})
