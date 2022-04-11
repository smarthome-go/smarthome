import { writable, Writable } from 'svelte/store';
import App from './App.svelte';

export interface reminder {
  id: number;
  name: string;
  description: string;
  priority: number;
  createdDate: number;
  dueDate: number;
  owner: string;
  userWasNotified: boolean;
  userWasNotifiedAt: number;
}

export const reminders: Writable<reminder[]> = writable([]);

export default new App({
  target: document.body,
})
