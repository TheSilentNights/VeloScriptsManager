import {useEffect, useState} from "react";
import {App, Button, Card, Select, Space, Spin, Typography} from "antd";
import {LoadingOutlined, ReloadOutlined, SaveOutlined} from "@ant-design/icons";
import {useConfigStore} from "../../store/configStore";
import {useScriptStore} from "../../store/scriptStore";
import {useEnvironmentStore} from "../../store/environmentStore";
import type {Script} from "../../types/models";
import type {ShortcutSlotPayload} from "../../ts/api";

export default function KeyboardShortCutPage() {
    const {message} = App.useApp();
    const font_size = useConfigStore((s) => s.font_size);
    const shortcuts = useConfigStore((s) => s.shortcuts);
    const configLoading = useConfigStore((s) => s.loading);
    const configError = useConfigStore((s) => s.error);
    const loadConfig = useConfigStore((s) => s.load);
    const saveConfig = useConfigStore((s) => s.update);
    const scripts = useScriptStore((s) => s.scripts);
    const loadScripts = useScriptStore((s) => s.load);
    const nameOf = useEnvironmentStore((s) => s.nameOf);
    const loadEnvironments = useEnvironmentStore((s) => s.load);

    //stores the temporary script commands and environments
    const [draft, setDraft] = useState<ShortcutSlotPayload[] | null>(null);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        void loadConfig();
        void loadScripts();
        void loadEnvironments();
    }, [loadConfig, loadScripts, loadEnvironments]);

    useEffect(() => {
        if (!configLoading && !configError) {
            setDraft(shortcuts);
        }
    }, [configLoading, configError, shortcuts]);

    useEffect(() => {
        if (configError) {
            message.error(`加载配置失败：${configError}`);
        }
    }, [configError, message]);

    const updateSlot = (index: number, slot: ShortcutSlotPayload) => {
        setDraft((prev) => {
            if (!prev) return prev;
            const next = [...prev];
            next[index] = slot;
            return next;
        });
    };

    const handleScriptChange = (index: number, value: string | undefined) => {
        const script = scripts.find((s) => s.id === value);
        updateSlot(index, {
            script_id: value ?? "",
            command: script ? [...script.command] : [],
            environments_id: script ? [...script.environments] : [],
        });
    };

    const handleRefresh = async () => {
        await Promise.all([loadConfig(), loadScripts(), loadEnvironments()]);
        if (useConfigStore.getState().error) return;
        message.success("已刷新配置");
    };

    const handleSave = async () => {
        if (draft === null || font_size === null) return;
        //stores the temporary draft of the shortcuts to be save
        setSaving(true);
        try {
            await saveConfig({font_size: font_size, shortcuts: draft});
            message.success("已保存快捷键槽位");
        } catch (e) {
            console.log(e);
            message.error(`保存失败：${(e as Error).message}`);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div style={pageContainerStyle}>
            <div style={cardStyle}>
                <div style={cardHeaderStyle}>
                    <Typography.Text strong style={{fontSize: 15}}>
                        键盘快捷键槽位
                    </Typography.Text>
                    <Space size={8}>
                        <Button
                            icon={<ReloadOutlined/>}
                            loading={configLoading}
                            disabled={saving}
                            onClick={handleRefresh}
                        >
                            刷新
                        </Button>
                        <Button
                            type="primary"
                            icon={<SaveOutlined/>}
                            loading={saving}
                            disabled={configLoading || draft === null}
                            onClick={handleSave}
                        >
                            保存
                        </Button>
                    </Space>
                </div>
                <Spin spinning={configLoading} indicator={<LoadingOutlined spin/>} size="small">
                    <div style={gridStyle}>
                        {(draft ?? []).map((slot, index) => (
                            <SlotCard
                                key={index}
                                index={index}
                                slot={slot}
                                scripts={scripts}
                                nameOf={nameOf}
                                onScriptChange={(value) => handleScriptChange(index, value)}
                                onCommandChange={(value) => updateSlot(index, {...slot, command: value})}
                                onEnvironmentsChange={(value) => updateSlot(index, {...slot, environments_id: value})}
                            />
                        ))}
                    </div>
                </Spin>
            </div>
        </div>
    );
}

interface SlotCardProps {
    index: number
    slot: ShortcutSlotPayload
    scripts: Script[]
    nameOf: (id: string) => string
    onScriptChange: (value: string | undefined) => void
    onCommandChange: (value: string[]) => void
    onEnvironmentsChange: (value: string[]) => void
}

function SlotCard({index, slot, scripts, nameOf, onScriptChange, onCommandChange, onEnvironmentsChange}: SlotCardProps) {
    const script = scripts.find((s) => s.id === slot.script_id);
    const scriptOptions = scripts.map((s) => ({
        label: s.name,
        value: s.id,
    }));
    if (slot.script_id && !script) {
        scriptOptions.push({label: "脚本已删除", value: slot.script_id});
    }

    return (
        <Card size="small" title={`槽位 ${index + 1}`}>
            <Space direction="vertical" style={{width: "100%"}} size={10}>
                <div>
                    <Typography.Text type="secondary" style={{fontSize: 12}}>
                        绑定脚本
                    </Typography.Text>
                    <Select
                        style={{width: "100%"}}
                        placeholder="未绑定"
                        value={slot.script_id || undefined}
                        allowClear
                        options={scriptOptions}
                        maxTagCount="responsive"
                        onChange={onScriptChange}
                    />
                </div>
                {script && (
                    <>
                        <div>
                            <Typography.Text type="secondary" style={{fontSize: 12}}>
                                启用命令 (command)
                            </Typography.Text>
                            <Select
                                mode="multiple"
                                style={{width: "100%"}}
                                placeholder="选择启用的命令节点"
                                value={slot.command}
                                options={script.command.map((c) => ({
                                    label: c,
                                    value: c,
                                }))}
                                maxTagCount="responsive"
                                onChange={onCommandChange}
                            />
                        </div>
                        <div>
                            <Typography.Text type="secondary" style={{fontSize: 12}}>
                                启用环境 (environments)
                            </Typography.Text>
                            <Select
                                mode="multiple"
                                style={{width: "100%"}}
                                placeholder="选择启用的环境"
                                value={slot.environments_id}
                                options={script.environments.map((id) => ({
                                    label: nameOf(id),
                                    value: id,
                                }))}
                                maxTagCount="responsive"
                                onChange={onEnvironmentsChange}
                            />
                        </div>
                    </>
                )}
                {slot.script_id && !script && (
                    <Typography.Text type="danger" style={{fontSize: 12}}>
                        绑定的脚本已被删除
                    </Typography.Text>
                )}
            </Space>
        </Card>
    );
}

const pageContainerStyle: React.CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    gap: 20,
    height: '100%',
};

const cardStyle: React.CSSProperties = {
    backgroundColor: '#ffffff',
    borderRadius: 12,
    boxShadow: '0 1px 3px rgba(0, 0, 0, 0.06)',
    padding: 24,
};

const cardHeaderStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
};

const gridStyle: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: 'repeat(5, minmax(0, 1fr))',
    gap: 12,
    marginTop: 16,
};
