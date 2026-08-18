export function SettingsPage() {
  return (
    <div style={pageContainerStyle}>
      <div style={cardStyle}>设置页面占位</div>
      <div style={cardStyle}>这里将展示应用偏好设置</div>
    </div>
  )
}

const pageContainerStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '2fr 1fr',
  gap: 20,
  height: '100%',
}

const cardStyle: React.CSSProperties = {
  backgroundColor: '#ffffff',
  borderRadius: 12,
  boxShadow: '0 1px 3px rgba(0, 0, 0, 0.06)',
  padding: 24,
}
