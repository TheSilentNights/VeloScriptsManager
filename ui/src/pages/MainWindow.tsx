import {useEffect} from "react";
import {App, ConfigProvider} from "antd";
import {AppShell} from "./AppShell.tsx";
import {useConfigStore} from "../store/configStore.ts";

export default function MainWindow() {
    const font_size = useConfigStore((s) => s.font_size);

    useEffect(() => {
        void useConfigStore.getState().load();
    }, []);

    return (
        <ConfigProvider theme={{token: {colorPrimary: "#4f46e5", fontSize: font_size ?? undefined}}}>
            <App style={{width: "100%", height: "100%"}}>
                <div style={mainWindowStyle}>
                    <AppShell/>
                </div>
            </App>
        </ConfigProvider>
    );
}

const mainWindowStyle: React.CSSProperties = {
    width: "100%",
    height: "100%",
    backgroundColor: "#ffffff",
};
