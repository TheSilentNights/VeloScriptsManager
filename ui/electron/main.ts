import { app, BrowserWindow, ipcMain } from 'electron'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import {getServerPort, startServer, stopServer} from "./launcher.ts"

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const isDev = !!process.env.VITE_DEV_SERVER_URL



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
    win.loadFile(path.join(__dirname, '../index.html'))
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
  if (!isDev){
    try {
      await startServer()
    } catch (e) {
      console.error(`start backend failed: ${(e as Error).message}`)
      app.quit()
      return
    }
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

ipcMain.handle('get-server-port', () => getServerPort())

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})

app.on('before-quit', () => {
  stopServer()
})
