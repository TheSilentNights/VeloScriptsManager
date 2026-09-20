import { type WebContents } from "electron";
import { globalShortcut } from "electron/main";

//format: ${key}_pressed
export function registerKeyInGlobalShortcut(webContents: WebContents, key: string) {
  globalShortcut.register(key, () => {
    webContents.send(key + "_pressed");
  });
}

export function unregisterKeyInGlobalShortcut(key: string) {
  globalShortcut.unregister(key);
}
