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
    let minuteDifference = minuteThen - minuteNow;
    let outputText = "In ";

    if (minuteDifference < 0) hourDifference--;

    if (hourDifference > 0) {
        outputText +=
            hourDifference > 1
                ? `${hourDifference} hours`
                : `${hourDifference} hour`;
    } else if (hourDifference < 0) {
        outputText +=
            hourDifference + 24 > 0
                ? `${hourDifference + 24} hours`
                : `${hourDifference + 24} hour`;
    }

    if (hourDifference !== 0 && minuteDifference !== 0)
        outputText += " and ";

    if (hourDifference > 0 && minuteDifference > 0) {
        outputText +=
            minuteDifference > 1
                ? `${minuteDifference} minutes`
                : `${minuteDifference} minute`;
    } else if (hourDifference === 0 && minuteDifference === 1) {
        outputText += ` ${60 - now.getSeconds()} seconds`;
    } else if (minuteDifference < 0) {
        outputText +=
            minuteDifference + 60 > 1
                ? `${minuteDifference + 60} minutes`
                : `${minuteDifference + 60} minute`;
    }
    return outputText
}

export const schedules: Writable<Schedule[]> = writable([])
export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
