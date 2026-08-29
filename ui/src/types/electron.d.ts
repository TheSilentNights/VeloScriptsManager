export {}

declare module "react" {
    interface CSSProperties {
        WebkitAppRegion?: "drag" | "no-drag"
    }
}

declare global {
  interface Window {
    electronAPI: {
      sendMessage: (msg: string) => void
      minimizeWindow: () => void
      maximizeWindow: () => void
      closeWindow: () => void
      getServerPort: () => Promise<number | null>
      onMaximizeChange: (callback: (isMaximized: boolean) => void) => void
    }
  }
}
