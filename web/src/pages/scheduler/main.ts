import type { homescript } from '../../homescript'
import { writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import App from './App.svelte'

export interface Schedule {
    id: number,
    owner: string,
    data: ScheduleData
}

export type ScheduleTargetMode = 'code' | 'hms' | 'devices'

export interface ScheduleData {
    name: string,
    hour: number,
    minute: number,
    targetMode: ScheduleTargetMode
    homescriptCode: string,
    homescriptTargetId: string,
    deviceJobs: DeviceJob[]
}

export interface DeviceJob {
    deviceId: string,
    powerOn: boolean
}


export interface SwitchResponse {
    id: string;
    name: string;
    powerOn: boolean;
    watts: number;
}

// Is used to calculate the time until the schedule's execution
// Returns a user-friendly string
export function timeUntilExecutionText(
    now: Date,
    hourThen: number,
    minuteThen: number
): string {
    now.setTime(now.getTime());
    const minuteNow = now.getMinutes();
    const hourNow = now.getHours();
    let hourDifference = hourThen - hourNow;
    const minuteDifference = minuteThen - minuteNow;
    let outputText = "In ";

    if (minuteDifference < 0) hourDifference--;

    if (hourDifference < 0)
        hourDifference += 24

    if (hourDifference > 0) {
        outputText +=
            hourDifference > 1
                ? `${hourDifference} hours`
                : `${hourDifference} hour`;
    }

    if (hourDifference !== 0 && minuteDifference !== 0)
        outputText += " and ";

    if (hourDifference === 0 && minuteDifference === 1) {
        outputText += ` ${60 - now.getSeconds()} seconds`;
    }
    else if (minuteDifference > 0) {
        outputText +=
            minuteDifference > 1
                ? `${minuteDifference} minutes`
                : `${minuteDifference} minute`;
    } else if (minuteDifference < 0) {
        outputText +=
            minuteDifference + 60 > 1
                ? `${minuteDifference + 60} minutes`
                : `${minuteDifference + 60} minute`;
    }
    return outputText
}

export const schedules: Writable<Schedule[]> = writable([])
export const schedulesLoaded: Writable<boolean> = writable(false);

// States that homescripts have been loaded
export const hmsLoaded: Writable<boolean> = writable(false)
export const homescripts: Writable<homescript[]> = writable([])

// States that devices have been loaded
export const devicesLoaded: Writable<boolean> = writable(false)
export const devices: Writable<SwitchResponse[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
