import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // El contrato OpenAPI declara `servers: - url: /api/v1` (mismo origen,
    // detrás de un reverse proxy en despliegue real). En desarrollo local
    // este proxy reproduce ese mismo origen para que shared/api use rutas
    // relativas y la cookie de sesión HttpOnly viaje sin fricción de CORS
    // ni de SameSite entre orígenes distintos (docs/06-api/estandar-openapi.md).
    proxy: {
      '/api': {
        target: process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
