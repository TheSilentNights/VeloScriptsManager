import {useEffect} from "react";
import {Button, Empty, Space, Spin, Typography} from "antd";
import {ReloadOutlined} from "@ant-design/icons";
import {useExecutionStore} from "../../store/executionStore";
import {ExecutionTile} from "./components/executions/ExecutionTile";

export function ExecutionsPage() {
    const executions = useExecutionStore((s) => s.executions);
    const loading = useExecutionStore((s) => s.loading);
    const load = useExecutionStore((s) => s.load);

    useEffect(() => {
        load();
    }, [load]);

    const sorted = [...executions].sort((a, b) => {
        const ar = a.status === "running" ? 0 : 1;
        const br = b.status === "running" ? 0 : 1;
        if (ar !== br) return ar - br;
        return (b.startedAt ?? "").localeCompare(a.startedAt ?? "");
    });

    const runningCount = executions.filter((e) => e.status === "running").length;

    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column",
                gap: 16,
                height: "100%",
                overflow: "auto",
            }}
        >
            <Space style={{justifyContent: "space-between"}}>
                <Typography.Text type="secondary">
                    {runningCount > 0
                        ? `当前有 ${runningCount} 个实例正在执行（运行中优先展示）`
                        : "当前没有正在执行的实例"}
                </Typography.Text>
                <Button icon={<ReloadOutlined/>} onClick={load} loading={loading}>
                    刷新
                </Button>
            </Space>
            <Spin spinning={loading && executions.length === 0}>
                {sorted.length === 0 ? (
                    <Empty description="暂无执行实例"/>
                ) : (
                    <div
                        style={{
                            display: "grid",
                            gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
                            gap: 16,
                            alignContent: "start",
                        }}
                    >
                        {sorted.map((e) => (
                            <ExecutionTile
                                key={e.executionId}
                                execution={e}
                            />
                        ))}
                    </div>
                )}
            </Spin>
        </div>
    );
}
