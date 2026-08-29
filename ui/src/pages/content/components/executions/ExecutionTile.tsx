import {App, Button, Card, Popconfirm, Tag, Typography, Space} from "antd";
import {StopOutlined} from "@ant-design/icons";
import type {ExecutionInfo} from "../../../../types/models";
import {useEnvironmentStore} from "../../../../store/environmentStore";
import {useExecutionStore} from "../../../../store/executionStore";

const STATUS_COLOR: Record<string, string> = {
    running: "green",
    finished: "default",
    failed: "red",
};

const STATUS_TEXT: Record<string, string> = {
    running: "执行中",
    finished: "已完成",
    failed: "失败",
};

export function ExecutionTile({
    execution,
}: {
    execution: ExecutionInfo
}) {
    const {message} = App.useApp();
    const nameOf = useEnvironmentStore((s) => s.nameOf);
    const remove = useExecutionStore((s) => s.remove);

    const handleKill = async () => {
        try {
            await remove(execution.executionId);
            message.success("已终止执行");
        } catch (e:any) {
            console.log(e)
            message.error(`终止失败：${e.response?.data.message || (e as Error).message}`);
        }
    };

    return (
        <Card
            size="small"
            title={
                <Typography.Text strong style={{fontSize: 15}}>
                    {execution.name}
                </Typography.Text>
            }
            extra={
                <Tag color={STATUS_COLOR[execution.status] ?? "default"}>
                    {STATUS_TEXT[execution.status] ?? execution.status}
                </Tag>
            }
        >
            <Space orientation="vertical" style={{width: "100%"}} size={10}>
                <div>
                    <Typography.Text type="secondary" style={{fontSize: 12}}>
                        启用的环境 (environments)
                    </Typography.Text>
                    <div>
                        {execution.environments.length === 0 ? (
                            <Typography.Text type="secondary">无</Typography.Text>
                        ) : (
                            <Space size={4} wrap>
                                {execution.environments.map((id) => (
                                    <Tag key={id}>{nameOf(id)}</Tag>
                                ))}
                            </Space>
                        )}
                    </div>
                </div>

                <div>
                    <Typography.Text type="secondary" style={{fontSize: 12}}>
                        命令 (command)
                    </Typography.Text>
                    <div>
                        {execution.command.length === 0 ? (
                            <Typography.Text type="secondary">无</Typography.Text>
                        ) : (
                            <Space size={4} wrap>
                                {execution.command.map((c, i) => (
                                    <Tag key={i} color="blue">
                                        {c}
                                    </Tag>
                                ))}
                            </Space>
                        )}
                    </div>
                </div>

                <Popconfirm
                    title="终止执行"
                    description="确定要终止该执行实例吗？"
                    okText="终止"
                    okButtonProps={{color: "danger"}}
                    cancelText="取消"
                    onConfirm={handleKill}
                >
                    <Button
                        danger
                        block
                        icon={<StopOutlined/>}
                    >
                        终止
                    </Button>
                </Popconfirm>
            </Space>
        </Card>
    );
}
