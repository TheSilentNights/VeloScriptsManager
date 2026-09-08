import {useEffect, useState} from "react";
import {App, Button, Empty, Space, Spin} from "antd";
import {LoadingOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import {useEventStore} from "../../store/eventStore";
import {useScriptStore} from "../../store/scriptStore";
import type {FileChangeEventPayload, TimeEventPayload} from "../../ts/api";
import {EventTile} from "./components/events/EventTile";
import {EventEditorModal} from "./components/events/EventEditorModal";

export function EventsPage() {
    const {message} = App.useApp();
    const events = useEventStore((s) => s.events);
    const loading = useEventStore((s) => s.loading);
    const load = useEventStore((s) => s.load);
    const registerFileChange = useEventStore((s) => s.registerFileChange);
    const registerTime = useEventStore((s) => s.registerTime);
    const scripts = useScriptStore((s) => s.scripts);
    const loadScripts = useScriptStore((s) => s.load);

    const [adding, setAdding] = useState(false);

    useEffect(() => {
        load();
        loadScripts();
    }, [load, loadScripts]);

    const handleRegister = async (payload: FileChangeEventPayload | TimeEventPayload) => {
        try {
            if ("path" in payload) {
                await registerFileChange(payload);
                const existsNow = useEventStore
                    .getState()
                    .events.some((item) => item.type === "fileChangeEvent" && item.eventId === payload.path);
                if (!existsNow) {
                    message.warning("注册可能未生效：监听路径无效时后端会静默失败，且该路径无法再次注册");
                    return;
                }
                message.success("文件监听事件注册成功");
            } else {
                await registerTime(payload);
                message.success("时间事件注册成功");
            }
            setAdding(false);
        } catch (e: any) {
            const serverMessage = e?.response?.data?.message;
            if (serverMessage) {
                message.error(`注册失败（HTTP ${e.response.status}）：${serverMessage}`);
            } else {
                message.error(`网络错误，请确认后端服务已启动：${(e as Error).message}`);
            }
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
                        onClick={() => setAdding(true)}
                    >
                        添加事件
                    </Button>
                </Space>
            </div>

            <Spin spinning={loading} indicator={<LoadingOutlined spin />} size={"large"}>
                {events.length === 0 && !loading ? (
                    <Empty description="暂无已注册事件"/>
                ) : (
                    <div style={gridStyle}>
                        {events.map((eventInfo) => (
                            <EventTile key={eventInfo.eventId} eventInfo={eventInfo}/>
                        ))}
                    </div>
                )}
            </Spin>

            <EventEditorModal
                open={adding}
                scripts={scripts}
                onCancel={() => setAdding(false)}
                onSubmit={handleRegister}
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
