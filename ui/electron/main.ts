import { app, BrowserWindow, ipcMain, Menu, Tray } from 'electron'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { getServerPort, startServer, stopServer } from "./launcher.ts"
import { registerKeyInGlobalShortcut, unregisterKeyInGlobalShortcut } from './keyboard.ts'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

let tray: Tray | null = null

const isDev = !!process.env.VITE_DEV_SERVER_URL

let mainWindow: BrowserWindow | null = null

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
    //load main.html
    win.loadURL(process.env.VITE_DEV_SERVER_URL + '/main/index.html')
    win.webContents.openDevTools()
  } else {
    win.loadFile(path.join(__dirname, '../renderer/main/index.html'))
  }
  mainWindow = win
}

function generatePopupContext(): Menu {
  return Menu.buildFromTemplate([
    {
      label: '退出',
      role: 'quit' // 使用内置角色，自动处理退出逻辑
    }
  ])
}


app.whenReady().then(async () => {
  tray = new Tray(path.join(__dirname, '../renderer/icon.png'))
  
  tray.on('click', () => {
    if (mainWindow) {
      mainWindow.show()
    }
  })

  tray.setContextMenu(generatePopupContext())


  if (!isDev) {
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

ipcMain.on('window-minimize-to-tray', () => {
  BrowserWindow.getFocusedWindow()?.hide()
})

ipcMain.on('window-minimize', () => {
  BrowserWindow.getFocusedWindow()?.minimize()
})

ipcMain.handle('window-maximize', async () => {
  const win = BrowserWindow.getFocusedWindow()
  if (!win) return false
  if (win.isMaximized()) {
    win.unmaximize()
    return false
  } else {
    win.maximize()
    return true
  }
})

ipcMain.on('register-key', (event, key) => {
  registerKeyInGlobalShortcut(event.sender, key)
})

ipcMain.on('unregister-key', (_event, key) => {
  unregisterKeyInGlobalShortcut(key)
})

ipcMain.on('window-close', () => {
  BrowserWindow.getFocusedWindow()?.close()
  tray?.destroy()
})

ipcMain.handle('get-server-port', () => getServerPort())

app.on('window-all-closed', () => {
  // prepare for tray menu
})

app.on('before-quit', () => {
  stopServer()
})

