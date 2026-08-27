import type {CSSProperties} from "react";
import {Button} from "antd";

export function ScriptsPage() {
  return (
    <div style={pageContainerStyle}>

    </div>
  )
}

const pageContainerStyle: CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '2fr 1fr',
  gap: 20,
  height: '100%',
}
