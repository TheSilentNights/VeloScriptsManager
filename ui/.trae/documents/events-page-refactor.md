# EventsPage 重构计划：事件列表 + 添加事件弹窗

## Context

后端已提供 `GET /api/v1/event/` 查询已注册事件（见 [API.md](file:///c:/develop/project/public/VeloScriptsManager/ui/API.md) 事件章节）。当前 [EventsPage.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/content/EventsPage.tsx) 是纯注册表单页，需重构为：

- 进入页面先展示已注册事件列表（网格卡片，与脚本页/执行记录页风格一致）
- 顶部提供「刷新」+「添加事件」按钮
- 「添加事件」打开 Modal：Segmented 选择文件监听/时间事件，脚本仅通过 Select 引用脚本 ID（不再提供 command/environmentsid 覆盖参数）
- 注册成功后关闭 Modal 并刷新列表
- 侧栏菜单「事件注册」改为「事件」（已与用户确认：网格卡片形式、菜单改名）

已知 API 限制：
- `GET /api/v1/event/` **不返回事件关联的脚本信息**，卡片只能展示类型 + 路径/间隔/重复
- 文件事件路径无效时后端**静默失败**（仍返回 200），需前端注册后校验列表给出 warning
- 无注销事件 API，卡片为纯只读
- `repeat` 当前后端仅触发一次（沿用现有 extra 提示文案）

## 修改文件清单

| 文件 | 操作 |
| --- | --- |
| [models.ts](file:///c:/develop/project/public/VeloScriptsManager/ui/src/types/models.ts) | 修改：新增 EventInfo 判别联合类型 |
| [api.ts](file:///c:/develop/project/public/VeloScriptsManager/ui/src/ts/api.ts) | 修改：新增 `fetchEvents()`；`ExecuteScriptPayload` 的 `command`/`environmentsid` 改可选 |
| `src/store/eventStore.ts` | 新建：zustand store（对照 executionStore） |
| `src/pages/content/components/events/EventTile.tsx` | 新建：事件卡片组件 |
| `src/pages/content/components/events/EventEditorModal.tsx` | 新建：添加事件 Modal（对照 ScriptEditorModal） |
| [EventsPage.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/content/EventsPage.tsx) | 重写：列表页骨架（对照 ScriptsPage） |
| [AppShell.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/AppShell.tsx) | 修改：菜单 label「事件注册」→「事件」（L36，key/图标不动） |

## 详细改动

### 1. src/types/models.ts

在 `ExecutionInfo` 之后新增（`event` 字段保持后端大写键名）：

```ts
export interface FileChangeEventDetail {
    ID: string
    Path: string
}

export interface TimeEventDetail {
    ID: string
    Interval: number
    Repeat: boolean
}

export interface FileChangeEventInfo {
    eventId: string
    type: "fileChangeEvent"
    event: FileChangeEventDetail
}

export interface TimeEventInfo {
    eventId: string
    type: "timeEvent"
    event: TimeEventDetail
}

export type EventInfo = FileChangeEventInfo | TimeEventInfo
```

消费方通过 `type === "fileChangeEvent"` 收窄后访问 `event.Path` / `event.Interval` / `event.Repeat`。

### 2. src/ts/api.ts

- import 增加 `EventInfo`
- `ExecuteScriptPayload` 改为（API.md：command/environmentsid 可选，不传时用脚本存储值；唯一消费方是被重写的 EventsPage，无连锁影响）：

```ts
export interface ExecuteScriptPayload {
    id: string
    command?: string[]
    environmentsid?: string[]
}
```

- 在 `deleteExecution` 之后、`ExecuteScriptPayload` 之前新增：

```ts
export function fetchEvents(): Promise<EventInfo[]> {
    return sendGet<EventInfo[]>("/api/v1/event/");
}
```

路径带尾部斜杠，与 `fetchScripts`/`getExecutions` 一致；`sendGet` 对 `[]` 放行，无需改动。

### 3. src/store/eventStore.ts（新建）

对照 [executionStore.ts](file:///c:/develop/project/public/VeloScriptsManager/ui/src/store/executionStore.ts)，注册动作对照 scriptStore 的 add 模式（API 成功后自动 `load()`）。不设 `remove`（后端无注销 API）：

```ts
interface EventState {
    events: EventInfo[]
    loading: boolean
    error: string | null
    load: () => Promise<void>
    registerFileChange: (payload: FileChangeEventPayload) => Promise<void>
    registerTime: (payload: TimeEventPayload) => Promise<void>
}
```

### 4. src/pages/content/components/events/EventTile.tsx（新建）

对照 [ExecutionTile.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/content/components/executions/ExecutionTile.tsx)：`Card size="small"` + `Typography.Text strong fontSize 15` 标题 + `Space orientation="vertical"`（antd 6 用 orientation）+ secondary fontSize 12 标签。纯展示，无 store 依赖、无操作按钮：

- 标题：类型图标 + 文本 —— 文件事件 `<FileSearchOutlined/> 文件监听事件`，时间事件 `<ClockCircleOutlined/> 时间事件`
- 文件事件 body：`监听路径 (path)` 标签 + `Typography.Text code` + `wordBreak: "break-all"` 显示 `event.Path`
- 时间事件 body：`触发间隔 (interval)` + `{event.Interval} 秒`；`重复触发 (repeat)` + `<Tag color={Repeat ? "blue" : "default"}>{Repeat ? "重复执行" : "仅触发一次"}</Tag>`
- 不展示脚本信息、eventId（接口不返回脚本；eventId 无操作用途）

### 5. src/pages/content/components/events/EventEditorModal.tsx（新建）

对照 [ScriptEditorModal.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/content/components/scripts/ScriptEditorModal.tsx) 的 props 与 Modal 属性（`destroyOnHidden`、`mask={{closable: false}}`、`layout="vertical"`、`preserve={false}`），另加 `confirmLoading={submitting}` 防双击：

```ts
interface EventEditorModalProps {
    open: boolean
    scripts: Script[]
    onCancel: () => void
    onSubmit: (payload: FileChangeEventPayload | TimeEventPayload) => Promise<void>
}
```

（4 个 props 解构按惯例拆多行）

- 内部状态：`eventType: "fileChange" | "time"`（默认 fileChange）、`submitting`
- `useEffect([open, form])`：`!open` 时跳过；打开时 `setEventType("fileChange")` + `form.resetFields()`
- 表单第一个元素为 Segmented（无 name）：沿用旧页 options（`FileSearchOutlined/文件监听事件`、`ClockCircleOutlined/时间事件`）
- 切换类型时 `form.resetFields(["path", "interval", "repeat"])`——只重置类型相关字段，scriptId 选择跨类型保留
- 条件字段（从旧 EventsPage 原样迁移，含 extra 文案）：
  - fileChange：`path` Input，required + whitespace 校验，`extra="路径无效时注册会失败，且该路径将无法再次注册"`
  - time：`interval` InputNumber（min 1、precision 0、width 100%）required；`repeat` Switch（valuePropName="checked"），`extra="当前后端实现仅触发一次"`，initialValues `repeat: false`
  - 公共：`scriptId` Select（options 来自传入的 `scripts`），required
- **不含** command/environmentsid 表单项
- `handleOk`：`submitting` 守卫 → `validateFields()` → 构造 `execute_script: {id: values.scriptId}` → 按 eventType 调 `onSubmit`（payload 联合类型）→ finally 复位 submitting；错误向上抛由父页处理（是否关 Modal 由父页决定）

### 6. src/pages/content/EventsPage.tsx（重写）

骨架照抄 [ScriptsPage.tsx](file:///c:/develop/project/public/VeloScriptsManager/ui/src/pages/content/ScriptsPage.tsx)：

- store：`useEventStore`（events/loading/load/registerFileChange/registerTime）+ `useScriptStore`（scripts/loadScripts）；**移除 environmentStore**（无环境覆盖参数）
- `useEffect`：`load()` + `loadScripts()`
- 顶部 `Space`（`justifyContent: "flex-end"`）：刷新按钮（`ReloadOutlined` + `loading={loading}` + `onClick={() => load()}`）+ 主按钮「添加事件」（`PlusOutlined` + `setAdding(true)`）
- 列表：`Spin`（`loading && events.length === 0` 时 Empty「暂无已注册事件」）→ 网格 `repeat(auto-fill, minmax(280px, 1fr))` + `events.map(e => <EventTile key={e.eventId} eventInfo={e}/>)`
- 底部挂 `<EventEditorModal open={adding} scripts={scripts} onCancel={() => setAdding(false)} onSubmit={handleRegister}/>`
- `handleRegister`（唯一业务函数，沿用旧页错误处理风格，`"path" in payload` 收窄分发）：

```ts
const handleRegister = async (payload: FileChangeEventPayload | TimeEventPayload) => {
    try {
        if ("path" in payload) {
            await registerFileChange(payload);
            const existsNow = useEventStore.getState().events.some(
                (item) => item.type === "fileChangeEvent" && item.eventId === payload.path,
            );
            if (!existsNow) {
                message.warning("注册可能未生效：监听路径无效时后端会静默失败，且该路径无法再次注册");
                return;
            }
            message.success("文件监听事件注册成功");
        } else {
            await registerTime(payload);
            message.success("时间事件注册成功");
        }
        setAdding(false);
    } catch (e: any) {
        const serverMessage = e?.response?.data?.message;
        if (serverMessage) {
            message.error(`注册失败（HTTP ${e.response.status}）：${serverMessage}`);
        } else {
            message.error(`网络错误，请确认后端服务已启动：${(e as Error).message}`);
        }
    }
};
```

说明：store 注册动作内部已 `load()`，await 返回后用 `useEventStore.getState()` 同步读取最新列表做静默失败校验（后端 200 但条目回滚时给 warning 且不关 Modal）；`eventId` 即文件监听路径，存在即注册成功（含同路径追加订阅者场景）。

- 旧文件底部样式 `pageContainerStyle`/`headerStyle`/`formCardStyle` 替换为 ScriptsPage 同款 `pageContainerStyle`/`gridStyle`

### 7. src/pages/AppShell.tsx

仅 L36：`label: "事件注册"` → `label: "事件"`。key `"events"`、图标、renderPage 逻辑不动。

## 实施顺序

1. `src/types/models.ts`（类型基座）
2. `src/ts/api.ts`（fetchEvents + payload 可选化）
3. `src/store/eventStore.ts`
4. `src/pages/content/components/events/EventTile.tsx`
5. `src/pages/content/components/events/EventEditorModal.tsx`
6. `src/pages/content/EventsPage.tsx` 重写
7. `src/pages/AppShell.tsx` label 修改

## 约束落实

- 全部新代码零注释；不触碰任何既有注释
- ≥4 参数的组件 props/函数参数列表拆为多行（仅 EventEditorModal 4 个 props 涉及）

## 验证（由用户执行）

- `npx tsc --noEmit` 类型检查
- 手动验收：侧栏显示「事件」→ 空列表 Empty → 刷新按钮 → 添加文件事件（卡片出现、路径显示）→ 添加时间事件（间隔/重复显示）→ 无效路径注册（warning 且 Modal 不关闭）→ 后端关闭时注册（网络错误提示）→ repeat=false 时间事件触发后刷新列表卡片消失属预期
