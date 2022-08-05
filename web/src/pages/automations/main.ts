import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface automation {
    id: number
    name: string
    description: string
    cronExpression: string
    cronDescription: string
    homescriptId: string
    owner: string
    enabled: boolean
    timingMode: 'normal' | 'sunrise' | 'sunset'
}

export interface addAutomation {
    name: string
    description: string
    hour: number
    minute: number
    days: number[]
    homescriptId: string
    enabled: boolean
    timingMode: 'normal' | 'sunrise' | 'sunset'
}


export interface homescript {
    owner: string
    data: {
        id: string
        name: string
        description: string
        quickActionsEnabled: boolean
        schedulerEnabled: boolean
        code: string
        mdIcon: string
    }
}

// Parses a valid cron-expression, if it is invalid, an error is thrown
export function parseCronExpressionToTime(expr: string): { hours: number, minutes: number, days: number[] } {
    if (expr === "* * * * *") return { days: [], hours: 0, minutes: 0 }
    const rawExpr = expr.split(" ")
    if (rawExpr.length != 5) throw Error(`Invalid cron-expression: '${expr}'`)
    // Days
    let days: number[] = []
    if (rawExpr[4] === "*")
        days = [0, 1, 2, 3, 4, 5, 6]
    else days = rawExpr[4].split(",").map(d => parseInt(d))
    return { hours: parseInt(rawExpr[1]), minutes: parseInt(rawExpr[0]), days: days }
}

// // Generates a cron-expression based on the provided data
// Logic ported from `backend: /core/scheduler/automation/cron.go`
export function generateCronExpression(hour: number, minute: number, days: number[]): string {
    let outputRep = ['', '', '*', '*', ''] // Cron-expression representation as list
    outputRep[0] = `${minute}` // Assign minute
    outputRep[1] = `${hour}` // Assign hour
    // Omit validation of days and time because the function is only used in a pre validated context
    if (days.length == 7) {
        // Set the days to '*' when all days are included in the list, does not check for duplicate days
        outputRep[4] = "*"
        return outputRep.join(" ")
    }
    // Append the days to the list
    for (let index = 0; index < days.length; index++) {
        outputRep[4] += `${days[index]}`
        // If the current day is not the last in the list, add a `
        if (index < days.length - 1) outputRep[4] += ","
    }
    return outputRep.join(" ")
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


// States that the automations have been loaded, is checked before displaying `no automations`
export const automationsLoaded: Writable<boolean> = writable(false)
export const automations: Writable<automation[]> = writable([])

// States that homescripts have been loaded
// used when trying to access the data of the automation's homescript
export const hmsLoaded: Writable<boolean> = writable(false)
export const homescripts: Writable<homescript[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
