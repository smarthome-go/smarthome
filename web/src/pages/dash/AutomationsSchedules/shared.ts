export interface automationWrapper {
    data: automation
    hours: number
    minutes: number
    days: number[]
}

type Trigger =
    | 'cron'
    | 'on_sunrise'
    | 'on_sunset'
    | 'interval'
    | 'on_login'
    | 'on_logout'
    | 'on_notification'
    | 'on_shutdown'

export interface automation {
    id: number
    name: string
    description: string
    homescriptId: string
    owner: string
    enabled: boolean
    disableOnce: boolean
    trigger: Trigger
    triggerCronExpression: string | null
    triggerInterval: number | null
    cronDescription: string | null
}

export interface Schedule {
    id: number
    owner: string
    data: ScheduleData
}

export interface ScheduleData {
    name: string
    hour: number
    minute: number
    targetMode: 'code' | 'hms' | 'switches'
    homescriptCode: string
    homescriptTargetId: string
    switchJobs: SwitchJob[]
}

export interface SwitchJob {
    switchId: string
    powerOn: boolean
}

// Is used to calculate the time until the schedule's execution
// Returns a user-friendly string
export function timeUntilExecutionText(
    now: Date,
    hourThen: number,
    minuteThen: number,
): string {
    now.setTime(now.getTime())
    const minuteNow = now.getMinutes()
    const hourNow = now.getHours()
    let hourDifference = hourThen - hourNow
    const minuteDifference = minuteThen - minuteNow
    let outputText = 'In '

    if (minuteDifference < 0) hourDifference--

    if (hourDifference < 0) hourDifference += 24

    if (hourDifference > 0) {
        outputText += hourDifference > 1
            ? `${hourDifference} h`
            : `${hourDifference} h`
    }

    if (hourDifference !== 0 && minuteDifference !== 0) outputText += ' and '

    if (hourDifference === 0 && minuteDifference === 1) {
        outputText += ` ${60 - now.getSeconds()} seconds`
    } else if (minuteDifference > 0) {
        outputText += minuteDifference > 1
            ? `${minuteDifference} min`
            : `${minuteDifference} min`
    } else if (minuteDifference < 0) {
        outputText += minuteDifference + 60 > 1
            ? `${minuteDifference + 60} min`
            : `${minuteDifference + 60} min`
    }
    return outputText
}
