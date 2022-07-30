import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface Schedule {
    id: number,
    owner: string,
    data: ScheduleData
}

export interface ScheduleData {
    name: string,
    hour: number,
    minute: number,
    targetMode: 'code' | 'hms' | 'switches'
    homescriptCode: string,
    homescriptTargetId: string,
    switchJobs: SwitchJob[]
}

export interface SwitchJob {
    switchId: string,
    powerOn: boolean
}

export const schedules: Writable<Schedule[]> = writable([])
export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
