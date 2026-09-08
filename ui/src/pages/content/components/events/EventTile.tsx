import {Card, Space, Tag, Typography} from "antd";
import {ClockCircleOutlined, FileSearchOutlined} from "@ant-design/icons";
import type {EventInfo} from "../../../../types/models";

export function EventTile({eventInfo}: {eventInfo: EventInfo}) {
    return (
        <Card
            size="small"
            title={
                eventInfo.type === "fileChangeEvent" ? (
                    <Typography.Text strong style={{fontSize: 15}}>
                        <FileSearchOutlined/> 文件监听事件
                    </Typography.Text>
                ) : (
                    <Typography.Text strong style={{fontSize: 15}}>
                        <ClockCircleOutlined/> 时间事件
                    </Typography.Text>
                )
            }
        >
            <Space orientation="vertical" style={{width: "100%"}} size={10}>
                {eventInfo.type === "fileChangeEvent" ? (
                    <div>
                        <Typography.Text type="secondary" style={{fontSize: 12}}>
                            监听路径 (path)
                        </Typography.Text>
                        <div>
                            <Typography.Text code style={{wordBreak: "break-all"}}>
                                {eventInfo.event.Path}
                            </Typography.Text>
                        </div>
                    </div>
                ) : (
                    <>
                        <div>
                            <Typography.Text type="secondary" style={{fontSize: 12}}>
                                触发间隔 (interval)
                            </Typography.Text>
                            <div>
                                <Typography.Text>{eventInfo.event.Interval} 秒</Typography.Text>
                            </div>
                        </div>
                        <div>
                            <Typography.Text type="secondary" style={{fontSize: 12}}>
                                重复触发 (repeat)
                            </Typography.Text>
                            <div>
                                <Tag color={eventInfo.event.Repeat ? "blue" : "default"}>
                                    {eventInfo.event.Repeat ? "重复执行" : "仅触发一次"}
                                </Tag>
                            </div>
                        </div>
                    </>
                )}
            </Space>
        </Card>
    );
}
