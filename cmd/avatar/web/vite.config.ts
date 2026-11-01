import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    // The content security policy of the page only allows files from the server
    assetsInlineLimit: 0,
  },
  server: {
    // The Go server answers the requests during development
    proxy: { '/api': 'http://localhost:8080' },
  },
  test: {
    environment: 'node',
  },
})
