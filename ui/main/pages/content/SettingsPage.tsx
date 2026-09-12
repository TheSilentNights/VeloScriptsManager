import {useEffect, useState} from "react";
import {App, Button, InputNumber, Spin, Typography} from "antd";
import {LoadingOutlined, ReloadOutlined, SaveOutlined} from "@ant-design/icons";
import {useConfigStore} from "../../store/configStore";

export function SettingsPage() {
    const {message} = App.useApp();
    const font_size = useConfigStore((s) => s.font_size);
    const configLoading = useConfigStore((s) => s.loading);
    const configError = useConfigStore((s) => s.error);
    const loadConfig = useConfigStore((s) => s.load);
    const saveConfig = useConfigStore((s) => s.update);
    const [fontSize, setFontSize] = useState<number | null>(null);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        void loadConfig();
    }, [loadConfig]);

    useEffect(() => {
        if (configError) {
            message.error(`加载设置失败：${configError}`);
        }
    }, [configError, message]);

    useEffect(() => {
        if (font_size !== null) {
            setFontSize(font_size);
        }
    }, [font_size]);

    const handleRefresh = async () => {
        await loadConfig();
        const state = useConfigStore.getState();
        if (state.error) return;
        setFontSize(state.font_size);
        message.success("已刷新配置");
    };

    const handleSave = async () => {
        if (fontSize === null) return;
        setSaving(true);
        try {
            await saveConfig({font_size: fontSize});
            message.success("已保存设置");
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
                <Typography.Text strong style={{fontSize: 15}}>
                    外观
                </Typography.Text>
                <Spin spinning={configLoading} indicator={<LoadingOutlined spin/>} size="small">
                    <div style={{display: "flex", alignItems: "center", gap: 12, marginTop: 16}}>
                        <Typography.Text>字体大小 (fontSize)</Typography.Text>
                        <InputNumber
                            min={12}
                            max={32}
                            step={1}
                            value={fontSize}
                            onChange={(value) => setFontSize(value)}
                            disabled={configLoading}
                        />
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
                            disabled={configLoading || fontSize === null}
                            onClick={handleSave}
                        >
                            保存
                        </Button>
                    </div>
                </Spin>
            </div>
        </div>
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
