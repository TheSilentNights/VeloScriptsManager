import {type CSSProperties} from 'react'
import {AppShell} from "./AppShell.tsx";

export default function MainWindow() {

    return (
        <div id={"main-window-background"} style={mainWindowBackgroundStyle}>
            <div id={"main-window"} style={{
                width: '100%',
                height: '100%',
                padding: '10px',
                boxSizing: 'border-box',
            }}>
                <AppShell/>
            </div>
        </div>
    )
}

const mainWindowBackgroundStyle: CSSProperties = {
    position: 'relative',
    width: '100%',
    height: '100%',
    borderRadius: '10px',
    backgroundColor: 'rgb(255 255 255 / 0.27)',
    boxShadow: '0 -4px 16px rgba(0, 0, 0, 0.06)',
    overflow: 'hidden',
}

