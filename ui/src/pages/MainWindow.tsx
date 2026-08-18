import { AppShell } from "./AppShell.tsx";

export default function MainWindow() {
  return (
    <div style={mainWindowStyle}>
      <AppShell />
    </div>
  )
}

const mainWindowStyle: React.CSSProperties = {
  width: "100%",
  height: "100%",
  backgroundColor: "#ffffff",
}
