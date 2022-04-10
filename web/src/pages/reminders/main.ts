import { writable, Writable } from 'svelte/store';
import App from './App.svelte';

export interface reminder {
  id: number;
  name: string;
  description: string;
  priority: number;
  createdDate: string;
  dueDate: string;
  owner: string;
  userWasNotified: boolean;
  userWasNotifiedAt: string;
}

export const reminders: Writable<reminder[]> = writable([]);

export default new App({
  target: document.body,
})
