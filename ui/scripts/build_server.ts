import { spawnSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import path from 'node:path'

const outDir = path.resolve('resources')
mkdirSync(outDir, { recursive: true })

console.log('Building Go server...')

const result = spawnSync(
  'go',
  ['build', '-ldflags', '-s -w', '-o', path.join(outDir, 'server.exe'), './server'],
  { cwd: '../service', stdio: 'inherit' },
)

if (result.status !== 0) {
  process.exit(result.status ?? 1)
}

console.log('Go server built successfully!')
