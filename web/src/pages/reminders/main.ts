import { writable } from 'svelte/store'
import type { Writable } from 'svelte/store'
import App from './App.svelte'

export interface reminder {
    id: number
    name: string
    description: string
    priority: number
    createdDate: number
    dueDate: number
    owner: string
    userWasNotified: boolean
    userWasNotifiedAt: number
}

export const reminders: Writable<reminder[]> = writable([])

export function sortReminders(input: reminder[]) {
    reminders.set(input.sort((a, b) => {
        // Sort by priority
        if (b.priority !== a.priority) {
            return b.priority - a.priority
        }
        // then sort by due date
        return a.dueDate - b.dueDate
    }))
}

export const loading: Writable<boolean> = writable(false)

export default new App({
    target: document.body,
})
