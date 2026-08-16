
import type { CSSProperties } from 'react'

export default function MainWindow(){
    return (
        <div id={"main-window"} style={mainWindowStyle}>
            <div style={glassBottomBarStyle} />
        </div>
    )
}

const mainWindowStyle: CSSProperties = {
    position: 'relative',
    width: '100%',
    height: '100%',
    borderRadius: '10px',
    backgroundColor: 'rgb(255 255 255 / 0.27)',
    boxShadow: '0 -4px 16px rgba(0, 0, 0, 0.06)',
    overflow: 'hidden',
}

const glassBottomBarStyle: CSSProperties = {
    position: 'absolute',
    bottom: 0,
}