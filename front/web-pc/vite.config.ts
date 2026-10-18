import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // 本地页面 localhost:5173；/api 由 Vite 转发到远程 Go API
  const defaultApiTarget = 'http://216.195.208.199:8886'
  const apiTarget =
    env.VITE_DEV_API_TARGET?.trim().replace(/\/+$/, '') || defaultApiTarget

  return {
    plugins: [
      vue(),
      {
        name: 'log-api-proxy',
        configureServer() {
          console.log(`[vite] /api proxy → ${apiTarget}`)
        },
      },
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@paper': fileURLToPath(new URL('.', import.meta.url)),
        '@ai-agent-paper/shared': fileURLToPath(
          new URL('./shared/src/index.ts', import.meta.url),
        ),
      },
    },
    server: {
      port: Number(env.VITE_DEV_PORT) || 5173,
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
