import { writable, Writable } from 'svelte/store'
import App from './App.svelte'
import "@fontsource/jetbrains-mono";

export interface LogEvent {
    id: number,
    name: string,
    description: string,
    // TRACE, DEBUG, INFO, WARN, ERROR, FATAL
    level: 0 | 1 | 2 | 3 | 4 | 5
    // Time as unix-millis
    time: number
}

export const logs: Writable<LogEvent[]> = writable([])



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
