import {useEffect, useState} from "react";
import {App, Button, Empty, Spin} from "antd";
import {PlusOutlined,LoadingOutlined} from "@ant-design/icons";
import {useScriptStore} from "../../store/scriptStore";
import {useEnvironmentStore} from "../../store/environmentStore";
import {ScriptTile} from "./components/scripts/ScriptTile";
import {ScriptEditorModal} from "./components/scripts/ScriptEditorModal";
import type {ScriptPayload} from "../../ts/api";

export function ScriptsPage() {
    const {message} = App.useApp();
    const scripts = useScriptStore((s) => s.scripts);
    const loading = useScriptStore((s) => s.loading);
    const scriptStore = useScriptStore((s) => s.load);
    const add = useScriptStore((s) => s.add);
    const environmentStore = useEnvironmentStore((s) => s.load);
    const environments = useEnvironmentStore((s) => s.environments);

    const [creating, setCreating] = useState(false);

    //update when scripts/environments changed
    useEffect(() => {
        scriptStore();
        environmentStore();
    }, [scriptStore, environmentStore]);

    const handleCreate = async (payload: ScriptPayload) => {
        try {
            await add(payload);
            message.success("已创建脚本");
            setCreating(false);
        } catch (e) {
            message.error(`创建失败：${(e as Error).message}`);
        }
    };

    return (
        <div style={pageContainerStyle}>
            <div style={{display: "flex", justifyContent: "flex-end"}}>
                <Button
                    type="primary"
                    icon={<PlusOutlined/>}
                    onClick={() => setCreating(true)}
                >
                    新建脚本
                </Button>
            </div>

            <Spin spinning={loading} indicator={<LoadingOutlined spin />} size={"large"}>
                {scripts.length === 0 && !loading ? (
                    <Empty description="暂无脚本"/>
                ) : (
                    <div style={gridStyle}>
                        {scripts.map((script) => (
                            <ScriptTile key={script.id} script={script}/>
                        ))}
                    </div>
                )}
            </Spin>

            <ScriptEditorModal
                open={creating}
                script={null}
                environments={environments}
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
