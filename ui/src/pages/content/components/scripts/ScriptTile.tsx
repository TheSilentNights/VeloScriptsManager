import {useState} from "react";
import {
    App,
    Button,
    Card,
    Popconfirm,
    Select,
    Space,
    Tag,
    Typography,
} from "antd";
import {
    DeleteOutlined,
    EditOutlined,
    PlayCircleOutlined,
} from "@ant-design/icons";
import type {Script} from "../../../../types/models";
import {executeScript, type ScriptPayload} from "../../../../ts/api";
import {useEnvironmentStore} from "../../../../store/environmentStore";
import {useScriptStore} from "../../../../store/scriptStore";
import {useExecutionStore, selectRunningCount} from "../../../../store/executionStore";
import {ScriptEditorModal} from "./ScriptEditorModal";

interface ScriptTileProps {
    script: Script
}

export function ScriptTile({script}: ScriptTileProps) {
    console.log(script)
    const syncKey = `${script.id}:${script.params.join(",")}:${script.environments.join(",")}`;

    return (
        <ScriptTileBody key={syncKey} script={script}/>
    );
}

function ScriptTileBody({script}: ScriptTileProps) {
    const {message} = App.useApp();
    const update = useScriptStore((s) => s.update);
    const remove = useScriptStore((s) => s.remove);
    const nameOf = useEnvironmentStore((s) => s.nameOf);
    const environments = useEnvironmentStore((s) => s.environments);
    const executions = useExecutionStore((s) => s.executions);
    const running = selectRunningCount(executions, script.id);

    const [editing, setEditing] = useState(false);
    const [executing, setExecuting] = useState(false);
    const [enabledParams, setEnabledParams] = useState<string[]>(script.params);
    const [enabledEnvironments, setEnabledEnvironments] = useState<string[]>(
        script.environments,
    );

    const handleExecute = async () => {
        setExecuting(true);
        try {
            const info = await executeScript(
                script.id,
                enabledParams,
                enabledEnvironments,
            );
            message.success(`已启动执行：${info.name}`);
        } catch (e:any) {
            console.log(e)
            message.error(`执行失败：${e.response?.data.message || (e as Error).message}`);
        } finally {
            setExecuting(false);
        }
    };

    const handleSave = async (payload: ScriptPayload) => {
        try {
            await update(script.id, payload);
            message.success("已保存修改");
            setEditing(false);
        } catch (e) {
            console.log(e)
            message.error(`保存失败：${(e as Error).message}`);
        }
    };

    const handleDelete = async () => {
        try {
            await remove(script.id);
            message.success("已删除脚本");
        } catch (e) {
            console.log(e)
            message.error(`删除失败：${(e as Error).message}`);
        }
    };

    return (
        <>
            <Card
                size="small"
                title={
                    <Typography.Text strong style={{fontSize: 15}}>
                        {script.name}
                    </Typography.Text>
                }
                extra={
                    <Space size={4}>
                        <Button
                            type="text"
                            size="small"
                            icon={<EditOutlined/>}
                            onClick={() => setEditing(true)}
                        />
                        <Popconfirm
                            title="删除脚本"
                            description="确定要删除该脚本吗？"
                            okText="删除"
                            okButtonProps={{color: "danger"}}
                            cancelText="取消"
                            onConfirm={handleDelete}
                        >
                            <Button
                                type="text"
                                size="small"
                                color="danger"
                                icon={<DeleteOutlined/>}
                            />
                        </Popconfirm>
                    </Space>
                }
            >
                <Space direction="vertical" style={{width: "100%"}} size={10}>
                    {running > 0 && (
                        <Typography.Text style={{color: "#52c41a", fontSize: 12}}>
                            <span
                                style={{
                                    display: "inline-block",
                                    width: 6,
                                    height: 6,
                                    borderRadius: "50%",
                                    backgroundColor: "#52c41a",
                                    marginRight: 6,
                                    verticalAlign: "middle",
                                }}
                            />
                            {running} instance{running > 1 ? "s" : ""} running
                        </Typography.Text>
                    )}
                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            workdir
                        </Typography.Text>
                        <div>
                            <Typography.Text code>{script.workDir}</Typography.Text>
                        </div>
                    </div>

                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            runner
                        </Typography.Text>
                        <div>
                            <Tag>{script.runner}</Tag>
                        </div>
                    </div>

                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            启用参数 (params)
                        </Typography.Text>
                        <Select
                            mode="multiple"
                            size="small"
                            style={{width: "100%"}}
                            placeholder="选择执行时启用的参数"
                            value={enabledParams}
                            onChange={setEnabledParams}
                            options={script.params.map((p) => ({
                                label: p,
                                value: p,
                            }))}
                            maxTagCount="responsive"
                        />
                    </div>

                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            启用环境 (environments)
                        </Typography.Text>
                        <Select
                            mode="multiple"
                            size="small"
                            style={{width: "100%"}}
                            placeholder="选择执行时启用的环境"
                            value={enabledEnvironments}
                            onChange={setEnabledEnvironments}
                            options={script.environments.map((id) => ({
                                label: nameOf(id),
                                value: id,
                            }))}
                            maxTagCount="responsive"
                        />
                    </div>

                    <Button
                        type="primary"
                        block
                        icon={<PlayCircleOutlined/>}
                        loading={executing}
                        onClick={handleExecute}
                    >
                        执行
                    </Button>
                </Space>
            </Card>

            <ScriptEditorModal
                open={editing}
                script={script}
                environments={environments}
                onCancel={() => setEditing(false)}
                onSubmit={handleSave}
            />
        </>
    );
}
