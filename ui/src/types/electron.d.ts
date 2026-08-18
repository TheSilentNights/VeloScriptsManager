export {}

declare global {
  interface Window {
    electronAPI: {
      sendMessage: (msg: string) => void
      minimizeWindow: () => void
      maximizeWindow: () => void
      closeWindow: () => void
      onMaximizeChange: (callback: (isMaximized: boolean) => void) => void
    }
  }
}
