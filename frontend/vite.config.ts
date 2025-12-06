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
      // Chat endpoints need longer timeout (AI responses can take 30+ seconds)
      '/api/v1/chat/message': {
        target: process.env.VITE_API_URL || (process.env.DOCKER ? 'http://backend:8080' : 'http://localhost:8080'),
        changeOrigin: true,
        secure: false,
        timeout: 120000, // 120 seconds for chat messages
      },
      // Other API endpoints
      '/api': {
        target: process.env.VITE_API_URL || (process.env.DOCKER ? 'http://backend:8080' : 'http://localhost:8080'),
        changeOrigin: true,
        secure: false,
        timeout: 60000, // 60 seconds for other API requests
      },
    },
  },
})
