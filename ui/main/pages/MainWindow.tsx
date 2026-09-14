import {useEffect} from "react";
import {App, ConfigProvider} from "antd";
import {AppShell} from "./AppShell.tsx";
import {useConfigStore} from "../store/configStore.ts";
import {executeScript} from "../ts/api.ts";

export default function MainWindow() {
    const font_size = useConfigStore((s) => s.font_size);

    useEffect(() => {
        void useConfigStore.getState().load();
    }, []);

    return (
        <ConfigProvider theme={{token: {colorPrimary: "#4f46e5", fontSize: font_size ?? undefined}}}>
            <App style={{width: "100%", height: "100%"}}>
                <ShortcutExecutor/>
                <div style={mainWindowStyle}>
                    <AppShell/>
                </div>
            </App>
        </ConfigProvider>
    );
}

function ShortcutExecutor() {
    const {message} = App.useApp();
    const shortcuts = useConfigStore((s) => s.shortcuts);
    const configLoading = useConfigStore((s) => s.loading);

    useEffect(() => {
        if (configLoading) return;
        const keys = [...new Set(shortcuts.map((slot) => slot.key).filter((key) => key !== ""))];
        const disposers = keys.map((key) => {
            window.electronAPI.registerKey(key);
            return window.electronAPI.onKeyPressed(key, () => {
                const slot = useConfigStore.getState().shortcuts.find((s) => s.key === key);
                if (!slot || slot.script_id === "") return;
                executeScript(slot.script_id, slot.command, slot.environments_id)
                    .then(() => {
                        message.success(`快捷键 ${key} 已触发脚本执行`);
                    })
                    .catch((e) => {
                        console.log(e);
                        message.error(`快捷键 ${key} 触发失败：${(e as Error).message}`);
                    });
            });
        });
        return () => {
            keys.forEach((key) => window.electronAPI.unregisterKey(key));
            disposers.forEach((dispose) => dispose());
        };
    }, [shortcuts, configLoading, message]);

    return null;
}

const mainWindowStyle: React.CSSProperties = {
    width: "100%",
    height: "100%",
    backgroundColor: "#ffffff",
};
