import {App, ConfigProvider} from "antd";
import {AppShell} from "./AppShell.tsx";

export default function MainWindow() {
    return (
        <ConfigProvider theme={{token: {colorPrimary: "#4f46e5"}}}>
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
