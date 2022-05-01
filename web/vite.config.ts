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
        editor: resolve(__dirname, 'html/editor.html'),
        404: resolve(__dirname, 'html/404.html'),
      }
    },
  },
})
