import {useEffect} from "react";
import {
    Button,
    Form,
    Input,
    Modal,
    Select,
    Space,
    Typography,
} from "antd";
import {MinusCircleOutlined, PlusOutlined} from "@ant-design/icons";
import type {Environment} from "../../../../types/models";
import type {EnvironmentPayload} from "../../../../ts/api";

interface EnvironmentEditorModalProps {
    open: boolean
    environment: Environment | null
    prefill?: Environment | null
    onCancel: () => void
    onSubmit: (payload: EnvironmentPayload) => Promise<void>
}

export function EnvironmentEditorModal({
    open,
    environment,
    prefill,
    onCancel,
    onSubmit,
}: EnvironmentEditorModalProps) {
    const [form] = Form.useForm();

    useEffect(() => {
        if (!open) return;
        if (environment) {
            form.resetFields();
            form.setFieldsValue({
                name: environment.name,
                paths: environment.paths ?? [],
                env: environment.env ?? [],
            });
        } else if (prefill) {
            form.resetFields();
            form.setFieldsValue({
                name: prefill.name,
                paths: prefill.paths ?? [],
                env: prefill.env ?? [],
            });
        } else {
            form.resetFields();
            form.setFieldsValue({env: [], paths: []});
        }
    }, [open, environment, prefill, form]);

    const handleOk = async () => {
        const values = await form.validateFields();
        await onSubmit({
            name: values.name,
            paths: values.paths ?? [],
            env: values.env ?? [],
        });
    };

    return (
        <Modal
            open={open}
            title={environment ? `编辑环境 - ${environment.name}` : "新建环境"}
            okText="保存"
            onOk={handleOk}
            onCancel={onCancel}
            destroyOnHidden={true}
            mask={{
                closable: false
            }}
        >
            <Form
                form={form}
                layout="vertical"
                preserve={false}
                initialValues={
                    environment
                        ? {
                              name: environment.name,
                              paths: environment.paths ?? [],
                              env: environment.env ?? [],
                          }
                        : prefill
                          ? {
                                name: prefill.name,
                                paths: prefill.paths ?? [],
                                env: prefill.env ?? [],
                            }
                          : {env: [], paths: []}
                }
            >
                <Form.Item
                    label="名称"
                    name="name"
                    rules={[{required: true, message: "请输入环境名称"}]}
                >
                    <Input placeholder="例如 java 21"/>
                </Form.Item>
                <Form.Item
                    label="关联路径 (paths)"
                    name="paths"
                    rules={[{required: true, message: "请至少输入一个路径"}]}
                    extra="执行时会拼接到系统 PATH 前面，例如 java 的 bin 目录"
                >
                    <Select
                        mode="tags"
                        placeholder="输入路径后回车添加"
                        tokenSeparators={[";"]}
                        open={false}
                    />
                </Form.Item>

                <Typography.Text type="secondary" style={{fontSize: 12}}>
                    环境变量 (env)
                </Typography.Text>
                <Form.List name="env">
                    {(fields, {add, remove}) => (
                        <>
                            {fields.map((field) => (
                                <Space
                                    key={field.key}
                                    style={{display: "flex", marginTop: 8}}
                                    align="center"
                                >
                                    <Form.Item
                                        name={[field.name, "key"]}
                                        rules={[{required: true, message: "请输入变量名"}]}
                                        style={{marginBottom: 0}}
                                    >
                                        <Input placeholder="变量名，例如 JAVA_HOME"/>
                                    </Form.Item>
                                    <Form.Item
                                        name={[field.name, "value"]}
                                        rules={[{required: true, message: "请输入变量值"}]}
                                        style={{marginBottom: 0}}
                                    >
                                        <Input placeholder="变量值"/>
                                    </Form.Item>
                                    <Button
                                        type="text"
                                        color="danger"
                                        icon={<MinusCircleOutlined/>}
                                        onClick={() => remove(field.name)}
                                    />
                                </Space>
                            ))}
                            <Form.Item style={{marginTop: 8, marginBottom: 0}}>
                                <Button
                                    type="dashed"
                                    block
                                    icon={<PlusOutlined/>}
                                    onClick={() => add()}
                                >
                                    添加变量
                                </Button>
                            </Form.Item>
                        </>
                    )}
                </Form.List>
            </Form>
        </Modal>
    );
}
