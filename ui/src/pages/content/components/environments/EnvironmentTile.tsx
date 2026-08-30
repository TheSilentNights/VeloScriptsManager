import {useState} from "react";
import {
    App,
    Button,
    Card,
    Popconfirm,
    Tag,
    Tooltip,
    Typography,
    Space,
} from "antd";
import {CopyOutlined, DeleteOutlined, EditOutlined} from "@ant-design/icons";
import type {Environment} from "../../../../types/models";
import {useEnvironmentStore} from "../../../../store/environmentStore";
import type {EnvironmentPayload} from "../../../../ts/api";
import {EnvironmentEditorModal} from "./EnvironmentEditorModal";

export function EnvironmentTile({
    environment,
}: {
    environment: Environment
}) {
    const {message} = App.useApp();
    const update = useEnvironmentStore((s) => s.update);
    const remove = useEnvironmentStore((s) => s.remove);
    const add = useEnvironmentStore((s) => s.add);
    const [editing, setEditing] = useState(false);
    const [copying, setCopying] = useState(false);

    const handleUpdate = async (payload: EnvironmentPayload) => {
        try {
            await update(environment.id, payload);
            await message.success("已保存修改");
        } catch (e) {
            console.log(e)
            await message.error(`保存失败：${(e as Error).message}`);
        }
    };

    const handleDelete = async () => {
        try {
            await remove(environment.id);
            await message.success("已删除环境");
        } catch (e) {
            console.log(e)
            await message.error(`删除失败：${(e as Error).message}`);
        }
    };

    const handleCopy = async (payload: EnvironmentPayload) => {
        try {
            await add(payload);
            await message.success("已创建环境");
            setCopying(false);
        } catch (e) {
            console.log(e)
            await message.error(`创建失败：${(e as Error).message}`);
        }
    };

    return (
        <>
            <Card
                size="small"
                title={
                    <Typography.Text strong style={{fontSize: 15}}>
                        {environment.name}
                    </Typography.Text>
                }
                extra={
                    <Space size={4}>
                        <Tooltip title="编辑环境">
                            <Button
                                type="text"
                                size="small"
                                icon={<EditOutlined/>}
                                onClick={() => setEditing(true)}
                            />
                        </Tooltip>
                        <Tooltip title="复制环境">
                            <Button
                                type="text"
                                size="small"
                                icon={<CopyOutlined/>}
                                onClick={() => setCopying(true)}
                            />
                        </Tooltip>
                        <Popconfirm
                            title="删除环境"
                            description="确定要删除该环境吗？"
                            okText="删除"
                            okButtonProps={{color: "danger"}}
                            cancelText="取消"
                            onConfirm={handleDelete}
                        >
                            <Tooltip title="删除环境">
                                <Button
                                    type="text"
                                    size="small"
                                    color="danger"
                                    icon={<DeleteOutlined/>}
                                />
                            </Tooltip>
                        </Popconfirm>
                    </Space>
                }
            >
                <Space orientation="vertical" style={{width: "100%"}} size={10}>
                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            关联路径 (paths)
                        </Typography.Text>
                        <div>
                            {environment.paths.length === 0 ? (
                                <Typography.Text type="secondary">无</Typography.Text>
                            ) : (
                                <Space size={4} wrap>
                                    {environment.paths.map((p) => (
                                        <Tag key={p}>{p}</Tag>
                                    ))}
                                </Space>
                            )}
                        </div>
                    </div>

                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            环境变量 (env)
                        </Typography.Text>
                        <div>
                            {environment.env.length === 0 ? (
                                <Typography.Text type="secondary">无</Typography.Text>
                            ) : (
                                <Space size={4} wrap>
                                    {environment.env.map((v) => (
                                        <Tag key={v.key} color="blue">
                                            {v.key}={v.value}
                                        </Tag>
                                    ))}
                                </Space>
                            )}
                        </div>
                    </div>
                </Space>
            </Card>

            <EnvironmentEditorModal
                open={editing}
                environment={environment}
                onCancel={() => setEditing(false)}
                onSubmit={async (payload: EnvironmentPayload) => {
                    await handleUpdate(payload);
                    setEditing(false);
                }}
            />

            <EnvironmentEditorModal
                open={copying}
                environment={null}
                prefill={environment}
                onCancel={() => setCopying(false)}
                onSubmit={handleCopy}
            />
        </>
    );
}
