import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/ws': {
        target: 'http://localhost:8080',
        ws: true,
      },
    },
  },
  resolve: {
    alias: {
      // Fix for @dagrejs/dagre ESM build issue
      '@dagrejs/dagre': resolve(__dirname, 'node_modules/@dagrejs/dagre/dist/dagre.cjs.js')
    }
  }
})
