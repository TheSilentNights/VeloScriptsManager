import { type CSSProperties, useState } from "react";
import { Layout, Menu, Typography, type MenuProps } from "antd";
import {
  SettingOutlined,
  CodeOutlined,
  DesktopOutlined,
  ApiOutlined,
} from "@ant-design/icons";
import "./AppShell.less";

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

export function AppShell() {
  const [page, setPage] = useState<PageKey>("scripts");

  return (
    <Layout style={{ height: "100%", background: "transparent" }}>
      <Sider width={230} className="glass-sider" style={glassSiderStyle}>
        <div style={{ padding: "20px 24px 8px" }}>
          <Typography.Title level={5} style={{ margin: 0, letterSpacing: 0.5 }}>
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
      <Layout style={{ background: "transparent" }}>
        <Content style={{ padding: 28 }}>
        </Content>
      </Layout>
    </Layout>
  );
}

const glassSiderStyle: CSSProperties = {
  backgroundColor: "#efefef",
  borderRadius: "10px",
  overflow: "hidden",
}
