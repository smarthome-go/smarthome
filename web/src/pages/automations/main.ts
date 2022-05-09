import { writable, Writable } from 'svelte/store'
import App from './App.svelte'

export interface automation {
	id:              number
	name:            string
	description:     string
	cronExpression:  string
	cronDescription: string
	homescriptId:    string
	owner:           string
	enabled:         boolean
	timingMode:      'normal' | 'sunrise' | 'sunset'
}

export const automations: Writable<automation[]> = writable([])

export const loading: Writable<boolean> = writable(false)

export default new App({
  target: document.body,
})
