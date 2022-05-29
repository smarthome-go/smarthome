import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface ScheduleData {
    id: number,
    name: string,
    owner: string,
    hour: number,
    minute: number,
    homescriptCode: string,
}

export const schedules: Writable<ScheduleData[]> = writable([])
export const loading: Writable<boolean> = writable(false)

export default new App({
	target: document.body,
})
