import { createServer } from 'vite'
import { context } from 'esbuild'
import { spawn, type ChildProcess } from 'node:child_process'
import electronPath from 'electron'

const viteServer = await createServer({
  configFile: 'vite.config.ts',
  server: {
    host: '127.0.0.1', 
  },
})
await viteServer.listen()

const address = viteServer.httpServer?.address()
const devUrl =
  typeof address === 'object' && address !== null
    ? `http://localhost:${address.port}`
    : 'http://localhost:5173'

let electronProcess: ChildProcess | null = null

function startElectron(): void {
  if (electronProcess) {
    electronProcess.kill()
  }
  
  electronProcess = spawn(electronPath as unknown as string, ['.'], {
    stdio: 'inherit',
    env: { ...process.env, VITE_DEV_SERVER_URL: devUrl },
  })
}

//build main
const mainContext = await context({
  entryPoints: ['electron/main.ts'],
  outfile: 'dist/electron/main.js',
  bundle: true,
  platform: 'node',
  format: 'esm',
  external: ['electron'],
  plugins: [
    {
      name: 'reload-electron',
      setup(build) {
        build.onEnd(() => startElectron())
      },
    },
  ],
})
await mainContext.watch()

// 4. 编译 Preload 脚本
const preloadContext = await context({
  entryPoints: ['electron/preload.ts'],
  outfile: 'dist/electron/preload.js',
  bundle: true,
  platform: 'node',
  format: 'cjs',
  external: ['electron'],
})
await preloadContext.watch()