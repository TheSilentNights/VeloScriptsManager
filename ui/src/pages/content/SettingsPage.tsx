import {useEffect, useState} from "react";
import {App, InputNumber, Spin, Typography} from "antd";
import {LoadingOutlined} from "@ant-design/icons";
import {fetchConfig, updateConfig} from "../../ts/api";

export function SettingsPage() {
    const {message} = App.useApp();
    const [fontSize, setFontSize] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);

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

    const handleFontSizeChange = async (value: number | null) => {
        if (value === null) return;
        const previous = fontSize;
        setFontSize(value);
        try {
            await updateConfig({fontSize: value});
        } catch (e) {
            console.log(e);
            setFontSize(previous);
            message.error(`保存失败：${(e as Error).message}`);
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
                            onChange={handleFontSizeChange}
                            disabled={loading}
                        />
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
