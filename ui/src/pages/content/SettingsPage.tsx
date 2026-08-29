import {useEffect, useState} from "react";
import {App, Button, InputNumber, Spin, Typography} from "antd";
import {LoadingOutlined, SaveOutlined} from "@ant-design/icons";
import {fetchConfig, updateConfig} from "../../ts/api";

export function SettingsPage() {
    const {message} = App.useApp();
    const [fontSize, setFontSize] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchConfig()
            .then((config) => {
                setFontSize(config.fontSize);
                setLoading(false);
            })
            .catch((e) => {
                console.log(e);
                message.error(`加载设置失败：${(e as Error).message}`);
                setLoading(false);
            });
    }, [message]);

    const handleSave = async () => {
        if (fontSize === null) return;
        setSaving(true);
        try {
            await updateConfig({fontSize: fontSize});
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
                <Spin spinning={loading} indicator={<LoadingOutlined spin/>} size="small">
                    <div style={{display: "flex", alignItems: "center", gap: 12, marginTop: 16}}>
                        <Typography.Text>字体大小 (fontSize)</Typography.Text>
                        <InputNumber
                            min={12}
                            max={32}
                            step={1}
                            value={fontSize}
                            onChange={(value) => setFontSize(value)}
                            disabled={loading}
                        />
                        <Button
                            type="primary"
                            icon={<SaveOutlined/>}
                            loading={saving}
                            disabled={loading || fontSize === null}
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
