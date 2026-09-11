import {useEffect, useState} from "react";
import {Form, Input, InputNumber, Modal, Segmented, Select, Switch} from "antd";
import {ClockCircleOutlined, FileSearchOutlined} from "@ant-design/icons";
import type {Script} from "../../../../types/models";
import type {FileChangeEventPayload, TimeEventPayload} from "../../../../ts/api";

type EventType = "fileChange" | "time";

interface EventEditorModalProps {
    open: boolean
    scripts: Script[]
    onCancel: () => void
    onSubmit: (payload: FileChangeEventPayload | TimeEventPayload) => Promise<void>
}

export function EventEditorModal({
    open,
    scripts,
    onCancel,
    onSubmit,
}: EventEditorModalProps) {
    const [form] = Form.useForm();
    const [eventType, setEventType] = useState<EventType>("fileChange");
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        if (!open) return;
        setEventType("fileChange");
        form.resetFields();
    }, [open, form]);

    const handleTypeChange = (value: EventType) => {
        setEventType(value);
        form.resetFields(["path", "interval", "repeat"]);
    };

    const handleOk = async () => {
        if (submitting) return;
        setSubmitting(true);
        try {
            const values = await form.validateFields();
            const executeScripts = values.scriptIds.map((id: string) => ({id: id}));
            if (eventType === "fileChange") {
                await onSubmit({
                    path: values.path!,
                    execute_scripts: executeScripts,
                });
            } else {
                await onSubmit({
                    interval: values.interval!,
                    repeat: values.repeat ?? false,
                    execute_scripts: executeScripts,
                });
            }
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <Modal
            open={open}
            title="添加事件"
            okText="注册"
            confirmLoading={submitting}
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
                initialValues={{repeat: false}}
            >
                <Form.Item>
                    <Segmented
                        value={eventType}
                        onChange={(value) => handleTypeChange(value as EventType)}
                        options={[
                            {value: "fileChange", icon: <FileSearchOutlined/>, label: "文件监听事件"},
                            {value: "time", icon: <ClockCircleOutlined/>, label: "时间事件"},
                        ]}
                    />
                </Form.Item>
                {eventType === "fileChange" && (
                    <Form.Item
                        label="监听路径 (path)"
                        name="path"
                        rules={[{required: true, whitespace: true, message: "请输入监听的文件或目录路径"}]}
                    >
                        <Input placeholder="例如 C:\watch\dir"/>
                    </Form.Item>
                )}
                {eventType === "time" && (
                    <>
                        <Form.Item
                            label="延迟秒数 (interval)"
                            name="interval"
                            rules={[{required: true, message: "请输入延迟秒数"}]}
                        >
                            <InputNumber min={1} precision={0} style={{width: "100%"}} placeholder="例如 60"/>
                        </Form.Item>
                        <Form.Item
                            label="重复触发 (repeat)"
                            name="repeat"
                            valuePropName="checked"
                            extra="当前后端实现仅触发一次"
                        >
                            <Switch/>
                        </Form.Item>
                    </>
                )}

                <Form.Item
                    label="触发脚本 (execute_scripts[].id)"
                    name="scriptIds"
                    rules={[{required: true, type: "array", message: "请选择事件触发时要执行的脚本"}]}
                >
                    <Select
                        mode="multiple"
                        placeholder="选择事件触发时要执行的脚本"
                        defaultActiveFirstOption={false}
                        options={scripts.map((s) => ({label: s.name, value: s.id}))}
                    />
                </Form.Item>
            </Form>
        </Modal>
    );
}
