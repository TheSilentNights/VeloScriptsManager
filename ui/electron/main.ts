import { app, BrowserWindow } from 'electron'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

function createWindow() {
  const win = new BrowserWindow({
    width: 900,
    height: 675,
    frame: false,           // 2. 隐藏原生标题栏和边框（透明窗口必须关闭原生外框）
    transparent: true,
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
}

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})