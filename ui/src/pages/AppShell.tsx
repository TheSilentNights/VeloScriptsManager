
import { type CSSProperties, useEffect, useState } from "react";
import {Layout, Menu, Typography, type MenuProps, Button} from "antd";
import {
  SettingOutlined,
  CodeOutlined,
  DesktopOutlined,
  ApiOutlined,
  MinusOutlined,
  BorderOutlined,
  FullscreenExitOutlined,
  CloseOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from "@ant-design/icons";
import "./AppShell.less";
import { ScriptsPage } from "./content/ScriptsPage.tsx";
import { AttachPage } from "./content/AttachPage.tsx";
import { EnvironmentsPage } from "./content/EnvironmentsPage.tsx";
import { SettingsPage } from "./content/SettingsPage.tsx";

const { Sider, Content } = Layout;

type PageKey = "environments" | "scripts" | "attach" | "settings";

type MenuItem = Required<MenuProps>["items"][number];

const menuItems: MenuItem[] = [
  {
    type: "group",
    label: "Operations",
    children: [
      { key: "scripts", icon: <CodeOutlined />, label: "脚本管理" },
      { key: "attach", icon: <DesktopOutlined />, label: "Attach" },
    ],
  },
  {
    type: "group",
    label: "Configuration",
    children: [
      { key: "environments", icon: <ApiOutlined />, label: "环境配置" },
      { key: "settings", icon: <SettingOutlined />, label: "设置" },
    ],
  },
];

const renderPage = (pageKey:PageKey) => {
  switch (pageKey) {
    case "scripts":
      return <ScriptsPage />
    case "attach":
      return <AttachPage />
    case "environments":
      return <EnvironmentsPage />
    case "settings":
      return <SettingsPage />
    default:
      return <ScriptsPage />
  }
}

export function AppShell() {
  const [page, setPage] = useState<PageKey>("scripts");

  const [collapsed, setCollapsed] = useState(false);

  function onCollapseChange(){
    setCollapsed(!collapsed)
  }

  return (
    <Layout style={{ height: "100%", background: "#f6f8fa" }}>
      <Sider width={230} className="glass-sider" style={siderStyle} trigger={null} collapsible collapsed={collapsed}>
        <div style={{ padding: "20px 24px 8px" }}>
          <Typography.Title level={5} style={{ margin: 0, letterSpacing: 0.5 }}>
            <img src={"src/assets/velo-scripts-icon-lightning-terminal.svg"} width={'20px'} alt={"ref"}/>
            VeloScriptsManager
          </Typography.Title>
        </div>
        <Menu
          mode="inline"
          items={menuItems}
          selectedKeys={[page]}
          onClick={(menuInfo) => setPage(menuInfo.key as PageKey)}
          style={{ background: "transparent", border: "none", marginTop: 8 }}
        />
      </Sider>
      <Layout style={contentLayoutStyle}>
        <Header collapsed={collapsed} onCollapseChange={onCollapseChange}/>
        <Content style={{ padding: 28, overflow: "auto" }}>
          {renderPage(page)}
        </Content>
      </Layout>
    </Layout>
  );
}

function Header({collapsed,onCollapseChange} : {collapsed:boolean,onCollapseChange:()=>void}){
  const [isMaximized, setIsMaximized] = useState(false);


  useEffect(() => {
    window.electronAPI.onMaximizeChange(setIsMaximized)
  }, [])

  return (
      <div style={headerStyle}>
        <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => onCollapseChange()}
            style={{
              fontSize: '16px',
              width: 16,
              height: 16,
              WebkitAppRegion: "no-drag",
            }}
        ></Button>
        <div style={windowControlsStyle}>
          <button
              type="button"
              style={windowButtonStyle}
              onClick={() => window.electronAPI.minimizeWindow()}
          >
            <MinusOutlined />
          </button>
          <button
              type="button"
              style={windowButtonStyle}
              onClick={() => window.electronAPI.maximizeWindow()}
          >
            {isMaximized ? <FullscreenExitOutlined /> : <BorderOutlined />}
          </button>
          <button
              type="button"
              style={closeButtonStyle}
              onClick={() => window.electronAPI.closeWindow()}
          >
            <CloseOutlined />
          </button>
        </div>
      </div>
  )
}

const siderStyle: CSSProperties = {
  backgroundColor: "#ffffff",
}

const contentLayoutStyle: CSSProperties = {
  backgroundColor: "#ffffff",
}

const headerStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  padding: "16px 24px",
  borderBottom: "1px solid #e1e4e8",
  WebkitAppRegion: "drag",
}

const windowControlsStyle: CSSProperties = {
  display: "flex",
  gap: 8,
  WebkitAppRegion: "no-drag",
}

const windowButtonStyle: CSSProperties = {
  width: 28,
  height: 28,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  border: "1px solid #d1d5da",
  borderRadius: 6,
  backgroundColor: "#ffffff",
  color: "#586069",
  cursor: "pointer",
}

const closeButtonStyle: CSSProperties = {
  width: 28,
  height: 28,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  border: "1px solid #d1d5da",
  borderRadius: 6,
  backgroundColor: "#ffffff",
  color: "#586069",
  cursor: "pointer",
}
