import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electronAPI', {
  sendMessage: (msg: string) => ipcRenderer.send('message', msg),
  minimizeWindow: () => ipcRenderer.send('window-minimize'),
  maximizeWindow: () => {
    return ipcRenderer.invoke('window-maximize')
  },
  minimizeToTray: () => ipcRenderer.send('window-minimize-to-tray'),
  closeWindow: () => ipcRenderer.send('window-close'),
  getServerPort: () => ipcRenderer.invoke('get-server-port'),
  onMaximizeChange: (callback: (isMaximized: boolean) => void) => {
    ipcRenderer.on('window-maximize-change', (_event, isMaximized) => callback(isMaximized))
  },
  registerKey: (key: string) => ipcRenderer.send('register-key', key),
  unregisterKey: (key: string) => ipcRenderer.send('unregister-key', key),
  onKeyPressed: (key: string, callback: () => void) => {
    const listener = () => callback()
    ipcRenderer.on(key + '_pressed', listener)
    return () => {
      ipcRenderer.removeListener(key + '_pressed', listener)
    }
  },
})