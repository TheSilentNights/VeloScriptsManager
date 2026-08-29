import { app, BrowserWindow, ipcMain } from 'electron'
import net from 'node:net'
import path from 'node:path'
import { spawn, type ChildProcess } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const isDev = !!process.env.VITE_DEV_SERVER_URL

let serverProcess: ChildProcess | null = null
let serverPort: number | null = null

function checkPortAvailable(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const server = net.createServer()
    server.once('error', () => resolve(false))
    server.once('listening', () => {
      server.close(() => resolve(true))
    })
    server.listen(port, '127.0.0.1')
  })
}

async function findAvailablePort(startPort: number): Promise<number> {
  for (let port = startPort; port < startPort + 100; port++) {
    if (await checkPortAvailable(port)) {
      return port
    }
  }
  throw new Error(`no available port in range ${startPort}-${startPort + 99}`)
}

function resolveServerCommand(): { command: string; args: string[]; cwd?: string } {
  if (isDev) {
    // dev：复用已编译的 server.exe（若存在），否则用 go run
    return {
      command: 'go',
      args: ['run', './server', '--port', String(serverPort)],
      cwd: path.resolve(__dirname, '../../service'),
    }
  }
  // prod：resources 目录下的 server.exe
  return {
    command: path.join(process.resourcesPath, 'server.exe'),
    args: ['--port', String(serverPort), '--release'],
  }
}

async function waitForServer(port: number, timeoutMs = 15000): Promise<void> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (serverProcess?.exitCode !== null && serverProcess?.exitCode !== undefined) {
      throw new Error(`server exited with code ${serverProcess.exitCode}`)
    }
    const ok = await new Promise<boolean>((resolve) => {
      const socket = net.connect(port, '127.0.0.1')
      socket.once('connect', () => {
        socket.destroy()
        resolve(true)
      })
      socket.once('error', () => resolve(false))
    })
    if (ok) return
    await new Promise((resolve) => setTimeout(resolve, 200))
  }
  throw new Error('server startup timeout')
}

async function startServer(): Promise<void> {
  serverPort = await findAvailablePort(19278)
  const { command, args, cwd } = resolveServerCommand()

  serverProcess = spawn(command, args, {
    cwd,
    stdio: 'inherit',
    windowsHide: true,
  })

  serverProcess.once('exit', (code) => {
    if (code !== 0 && code !== null) {
      console.error(`backend server exited with code ${code}`)
    }
  })

  await waitForServer(serverPort)
}

function stopServer(): void {
  if (!serverProcess || serverProcess.killed) return
  serverProcess.kill()
  serverProcess = null
}

function createWindow() {
  const win = new BrowserWindow({
    width: 900,
    height: 675,
    frame: false,           // 2. 隐藏原生标题栏和边框（透明窗口必须关闭原生外框）
    backgroundColor: '#ffffff',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
    },
  })

  // 根据 dev 环境变量选择加载地址
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    win.loadFile(path.join(__dirname, '../renderer/index.html'))
  }

  win.webContents.openDevTools()

  win.on('maximize', () => {
    win.webContents.send('window-maximize-change', true)
  })

  win.on('unmaximize', () => {
    win.webContents.send('window-maximize-change', false)
  })
}

app.whenReady().then(async () => {
  try {
    await startServer()
  } catch (e) {
    console.error(`start backend failed: ${(e as Error).message}`)
    app.quit()
    return
  }
  createWindow()
})

ipcMain.on('window-minimize', () => {
  BrowserWindow.getFocusedWindow()?.minimize()
})

ipcMain.on('window-maximize', () => {
  const win = BrowserWindow.getFocusedWindow()
  if (!win) return
  if (win.isMaximized()) {
    win.unmaximize()
  } else {
    win.maximize()
  }
})

ipcMain.on('window-close', () => {
  BrowserWindow.getFocusedWindow()?.close()
})

ipcMain.handle('get-server-port', () => serverPort)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})

app.on('before-quit', () => {
  stopServer()
})
