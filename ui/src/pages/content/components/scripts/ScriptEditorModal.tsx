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
    environments: Environment[]
    onCancel: () => void
    onSubmit: (payload: ScriptPayload) => Promise<void>
}

export function ScriptEditorModal({
    open,
    script,
    environments,
    onCancel,
    onSubmit,
}: ScriptEditorModalProps) {
    const [form] = Form.useForm();
    const [paramSearch, setParamSearch] = useState("");

    const addParamFromSearch = () => {
        const value = paramSearch.trim();
        if (!value) return;
        const current: string[] = form.getFieldValue("params") ?? [];
        if (!current.includes(value)) {
            form.setFieldsValue({params: [...current, value]});
        }
        setParamSearch("");
    };

    useEffect(() => {
        if (!open) return;
        if (script) {
            form.setFieldsValue({
                name: script.name,
                workDir: script.workDir,
                runner: script.runner,
                params: script.params ?? [],
                environments: script.environments ?? [],
            });
        } else {
            form.resetFields();
            form.setFieldsValue({params: [], environments: []});
        }
    }, [open, script, form]);

    const handleOk = async () => {
        const values = await form.validateFields();
        await onSubmit({
            name: values.name,
            workDir: values.workDir,
            runner: values.runner,
            params: values.params ?? [],
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
                              runner: script.runner,
                              params: script.params ?? [],
                              environments: script.environments ?? [],
                          }
                        : {params: [], environments: []}
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
                <Form.Item
                    label="执行程序 (runner)"
                    name="runner"
                    rules={[{required: true, message: "请输入执行程序"}]}
                >
                    <Input placeholder="例如 npm"/>
                </Form.Item>
                <Form.Item label="参数 (params)" name="params">
                    <Select
                        mode="tags"
                        placeholder="输入参数后回车添加"
                        tokenSeparators={[",", " "]}
                        onChange={() => setParamSearch("")}
                        showSearch={{
                            onSearch: setParamSearch,
                            searchValue: paramSearch
                        }}
                        popupRender={(menu) => (
                            <>
                                {menu}
                                <Divider style={{margin: "4px 0"}}/>
                                <Button
                                    type="text"
                                    icon={<PlusOutlined/>}
                                    block
                                    disabled={!paramSearch.trim()}
                                    onClick={addParamFromSearch}
                                >
                                    添加 “{paramSearch.trim() || "参数"}”
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
