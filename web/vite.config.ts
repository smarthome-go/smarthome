import { svelte } from '@sveltejs/vite-plugin-svelte'
import { resolve } from 'path'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    rollupOptions: {
      input: {
        login: resolve(__dirname, 'html/login.html'),
        dash: resolve(__dirname, 'html/dash.html'),
        rooms: resolve(__dirname, 'html/rooms.html'),
        reminders: resolve(__dirname, 'html/reminders.html'),
        profile: resolve(__dirname, 'html/profile.html'),
        users: resolve(__dirname, 'html/users.html'),
        hmsEditor: resolve(__dirname, 'html/hmsEditor.html'),
        homescript: resolve(__dirname, 'html/homescript.html'),
        automations: resolve(__dirname, 'html/automations.html'),
        scheduler: resolve(__dirname, 'html/scheduler.html'),
        404: resolve(__dirname, 'html/404.html'),
      }
    },
  },
})
