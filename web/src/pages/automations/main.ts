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
	}
}

export const automations: Writable<automation[]> = writable([])

export const homescripts: Writable<homescript[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
	target: document.body,
})
