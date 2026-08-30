import {useEffect, useState} from "react";
import {App, Button, Empty, Space, Spin} from "antd";
import {PlusOutlined, LoadingOutlined, ReloadOutlined} from "@ant-design/icons";
import {useEnvironmentStore} from "../../store/environmentStore";
import type {EnvironmentPayload} from "../../ts/api";
import {EnvironmentTile} from "./components/environments/EnvironmentTile";
import {EnvironmentEditorModal} from "./components/environments/EnvironmentEditorModal";

export function EnvironmentsPage() {
    const {message} = App.useApp();
    const environments = useEnvironmentStore((s) => s.environments);
    const loading = useEnvironmentStore((s) => s.loading);
    const load = useEnvironmentStore((s) => s.load);
    const add = useEnvironmentStore((s) => s.add);

    const [creating, setCreating] = useState(false);

    useEffect(() => {
        load();
    }, [load]);

    const handleCreate = async (payload: EnvironmentPayload) => {
        try {
            await add(payload);
            message.success("已创建环境");
            setCreating(false);
        } catch (e) {
            message.error(`创建失败：${(e as Error).message}`);
        }
    };

    return (
        <div style={pageContainerStyle}>
            <div style={{display: "flex", justifyContent: "flex-end"}}>
                <Space>
                    <Button
                        icon={<ReloadOutlined/>}
                        loading={loading}
                        onClick={() => load()}
                    >
                        刷新
                    </Button>
                    <Button
                        type="primary"
                        icon={<PlusOutlined/>}
                        onClick={() => setCreating(true)}
                    >
                        新建环境
                    </Button>
                </Space>
            </div>

            <Spin spinning={loading} indicator={<LoadingOutlined spin/>} size="large">
                {environments.length === 0 && !loading ? (
                    <Empty description="暂无环境"/>
                ) : (
                    <div style={gridStyle}>
                        {environments.map((e) => (
                            <EnvironmentTile key={e.id} environment={e}/>
                        ))}
                    </div>
                )}
            </Spin>

            <EnvironmentEditorModal
                open={creating}
                environment={null}
                onCancel={() => setCreating(false)}
                onSubmit={handleCreate}
            />
        </div>
    );
}

const pageContainerStyle: React.CSSProperties = {
    display: "flex",
    flexDirection: "column",
    gap: 16,
    height: "100%",
};

const gridStyle: React.CSSProperties = {
    display: "grid",
    gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
    gap: 16,
    alignContent: "start",
};
