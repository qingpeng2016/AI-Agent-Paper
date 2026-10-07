import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@paper': fileURLToPath(new URL('../../../agent/paper', import.meta.url)),
      // agent/paper 在 web-pc 外，需显式指向 workspace shared（否则无法解析 node 包名）
      '@ai-token-mall/shared': fileURLToPath(
        new URL('../shared/src/index.ts', import.meta.url),
      ),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8886',
        changeOrigin: true,
      },
    },
  },
})
