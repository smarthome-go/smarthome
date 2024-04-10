import { writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import App from './App.svelte'
import "@fontsource/jetbrains-mono";

export interface systemConfig {
    automationEnabled: boolean,
    lockDownMode: boolean,
    openWeatherMapApiKey: string,
    latitude: number,
    longitude: number
    mqtt: mqttSystemConfig
}

export interface mqttSystemConfig {
    enabled: boolean,
    host: string,
    port: number,
    username: string,
    password: string,
}

export interface mqttStatus {
    working: boolean,
    error: string | null
}

export interface logEvent {
    id: number,
    name: string,
    description: string,
    // TRACE, DEBUG, INFO, WARN, ERROR, FATAL
    level: 0 | 1 | 2 | 3 | 4 | 5
    // Time as unix-millis
    time: number
}

export const logs: Writable<logEvent[]> = writable([])

export const levels = [
    { label: "TRACE", color: "var(--clr-priority-low)" },
    { label: "DEBUG", color: "var(--clr-priority-low)" },
    { label: "INFO", color: "var(--clr-success)" },
    { label: "WARN", color: "var(--clr-warn) " },
    { label: "ERROR", color: "var(--clr-error)" },
    { label: "FATAL", color: "var(--clr-priority-medium)" },
];

export default new App({
    target: document.body,
})
