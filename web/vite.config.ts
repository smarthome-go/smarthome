import { svelte } from '@sveltejs/vite-plugin-svelte'
import { resolve } from 'path'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [svelte()],
    build: {
        chunkSizeWarningLimit: 515,
        rollupOptions: {
            input: {
                // Login / logout pages
                login: resolve(__dirname, 'html/login.html'),
                404: resolve(__dirname, 'html/404.html'),
                // User apps
                dash: resolve(__dirname, 'html/dash.html'),
                rooms: resolve(__dirname, 'html/rooms.html'),
                reminders: resolve(__dirname, 'html/reminders.html'),
                scheduler: resolve(__dirname, 'html/scheduler.html'),
                automations: resolve(__dirname, 'html/automations.html'),
                homescript: resolve(__dirname, 'html/homescript.html'),
                profile: resolve(__dirname, 'html/profile.html'),
                // Admin apps
                users: resolve(__dirname, 'html/users.html'),
                system: resolve(__dirname, 'html/system.html'),
                // Hidden pages
                hmsEditor: resolve(__dirname, 'html/hmsEditor.html'),
            },
            output: {
                manualChunks: (id: any) => {
                    if (id.includes("node_modules")) {
                        if (id.includes("@smui") || id.includes('@material')) {
                            return "vendor_mui";
                        } else if (id.includes("@lezer") || id.includes("@codemirror")) {
                            return "vendor_codemirror"
                        } else if (id.includes("chart.js") || id.includes("chartjs-adapter-date-fns") || id.includes("date-fns")) {
                            return "vendor_chartjs"
                        }
                        return "vendor"; // Remaining chunks end up here
                    }
                },
            }
        },
    },
})
