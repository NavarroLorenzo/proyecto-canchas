import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// Forzar carga del .env al momento del build
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [react()],
    define: {
      'process.env': env
    },
    server: {
      host: true,
      port: 5173,
    },
  }
})
