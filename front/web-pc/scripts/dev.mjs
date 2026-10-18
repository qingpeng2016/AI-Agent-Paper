/**
 * pnpm dev [host:port | http://host:port] [-- vite flags...]
 * 第一个 host:port 写入 VITE_DEV_API_TARGET；无则 vite.config 默认后端。
 */
import { spawnSync } from 'node:child_process'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), '..')

function parseApiTargetArg(raw) {
  const t = String(raw ?? '').trim()
  if (!t || t.startsWith('-')) return ''
  if (/^https?:\/\//i.test(t)) return t.replace(/\/+$/, '')
  if (/^[^:]+:\d+$/.test(t)) return `http://${t}`
  return ''
}

const argv = process.argv.slice(2)
const dashIdx = argv.indexOf('--')
const beforeDash = dashIdx >= 0 ? argv.slice(0, dashIdx) : argv
const viteArgs =
  dashIdx >= 0 ? argv.slice(dashIdx + 1) : argv.filter((a) => a.startsWith('-'))

const apiArg = beforeDash.find((a) => !a.startsWith('-'))
const apiTarget = parseApiTargetArg(apiArg)
const env = { ...process.env }
if (apiTarget) {
  env.VITE_DEV_API_TARGET = apiTarget
  console.log(`[dev] /api → ${apiTarget}`)
}

const result = spawnSync(process.platform === 'win32' ? 'vite.cmd' : 'vite', viteArgs, {
  cwd: root,
  env,
  stdio: 'inherit',
  shell: process.platform === 'win32',
})

process.exit(result.status ?? 1)
