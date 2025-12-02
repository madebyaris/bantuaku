import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    host: true,
    watch: {
      usePolling: true,
      interval: 1000,
    },
    proxy: {
      '/api': {
        // In Docker, use service name 'backend', otherwise use localhost
        target: process.env.VITE_API_URL || (process.env.DOCKER ? 'http://backend:8080' : 'http://localhost:8080'),
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
