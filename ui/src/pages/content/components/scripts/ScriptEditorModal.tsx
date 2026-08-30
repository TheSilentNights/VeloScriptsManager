import {useEffect, useState} from "react";
import {
    Button,
    Divider,
    Form,
    Input,
    Modal,
    Select,
} from "antd";
import {PlusOutlined} from "@ant-design/icons";
import type {Environment, Script} from "../../../../types/models";
import type {ScriptPayload} from "../../../../ts/api";

interface ScriptEditorModalProps {
    open: boolean
    script: Script | null
    prefill?: Script | null
    environments: Environment[]
    onCancel: () => void
    onSubmit: (payload: ScriptPayload) => Promise<void>
}

export function ScriptEditorModal({
    open,
    script,
    prefill,
    environments,
    onCancel,
    onSubmit,
}: ScriptEditorModalProps) {
    const [form] = Form.useForm();
    const [commandSearch, setCommandSearch] = useState("");

    const addCommandFromSearch = () => {
        const value = commandSearch.trim();
        if (!value) return;
        const current: string[] = form.getFieldValue("command") ?? [];
        if (!current.includes(value)) {
            form.setFieldsValue({command: [...current, value]});
        }
        setCommandSearch("");
    };

    useEffect(() => {
        if (!open) return;
        if (script) {
            form.setFieldsValue({
                name: script.name,
                workDir: script.workDir,
                command: script.command ?? [],
                environments: script.environments ?? [],
            });
        } else if (prefill) {
            form.setFieldsValue({
                name: prefill.name,
                workDir: prefill.workDir,
                command: prefill.command ?? [],
                environments: prefill.environments ?? [],
            });
        } else {
            form.resetFields();
            form.setFieldsValue({command: [], environments: []});
        }
    }, [open, script, prefill, form]);

    const handleOk = async () => {
        const values = await form.validateFields();
        await onSubmit({
            name: values.name,
            workDir: values.workDir,
            command: values.command ?? [],
            environmentsId: values.environments ?? [],
        });
    };

    return (
        <Modal
            open={open}
            title={script ? `编辑脚本 - ${script.name}` : "新建脚本"}
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
                    script
                        ? {
                              name: script.name,
                              workDir: script.workDir,
                              command: script.command ?? [],
                              environments: script.environments ?? [],
                          }
                        : prefill
                          ? {
                                name: prefill.name,
                                workDir: prefill.workDir,
                                command: prefill.command ?? [],
                                environments: prefill.environments ?? [],
                            }
                          : {command: [], environments: []}
                }
            >
                <Form.Item
                    label="名称"
                    name="name"
                    rules={[{required: true, message: "请输入脚本名称"}]}
                >
                    <Input placeholder="例如 build"/>
                </Form.Item>
                <Form.Item
                    label="工作目录 (workdir)"
                    name="workDir"
                    rules={[{required: true, message: "请输入工作目录"}]}
                >
                    <Input placeholder="例如 C:\\repo"/>
                </Form.Item>
                <Form.Item label="命令 (command)" name="command">
                    <Select
                        mode="tags"
                        placeholder="输入命令节点后回车添加"
                        tokenSeparators={[",", " "]}
                        onChange={() => setCommandSearch("")}
                        showSearch={{
                            onSearch: setCommandSearch,
                            searchValue: commandSearch
                        }}
                        popupRender={(menu) => (
                            <>
                                {menu}
                                <Divider style={{margin: "4px 0"}}/>
                                <Button
                                    type="text"
                                    icon={<PlusOutlined/>}
                                    block
                                    disabled={!commandSearch.trim()}
                                    onClick={addCommandFromSearch}
                                >
                                    添加 “{commandSearch.trim() || "命令"}”
                                </Button>
                            </>
                        )}
                    />
                </Form.Item>
                <Form.Item label="绑定环境 (environments)" name="environments">
                    <Select
                        mode="multiple"
                        placeholder="选择要应用的环境"
                        options={environments.map((e) => ({
                            label: e.name,
                            value: e.id,
                        }))}
                    />
                </Form.Item>
            </Form>
        </Modal>
    );
}
