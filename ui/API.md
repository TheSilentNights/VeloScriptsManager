# VeloScriptsManager 后端 API 文档

> 依据 `server/router.go`、`server/routers/`、`models/` 与相关 service/executor 实现整理。

## 通用约定

- **Base URL**：`http://127.0.0.1:19278`（默认端口 `19278`，可用启动参数 `--port` 修改；`--release` 切换为生产模式）。
- **响应信封**：所有接口（`GET /status` 除外）返回 `{ "message": "...", "data": ... }`，部分接口成功时不带 `data`。结果状态由 **HTTP 状态码** 表达（200/400/404/500），信封内没有 `code` 字段。
- **请求绑定**：请求体同时支持 `application/json` 与 form（`x-www-form-urlencoded` / `multipart`）。form 与 json 的字段名可能不同，见各接口表格。
- **CORS**：允许来源 `http://localhost:5173`、`http://127.0.0.1:5173`、任意 `http://127.0.0.1:*` 端口及 `origin: null`；允许方法 `GET`、`POST`、`PUT`、`DELETE`、`OPTIONS`；允许携带凭据。
- **时间格式**：`startedAt` 等时间字段为 RFC3339（如 `2026-09-08T12:00:00+08:00`）。

## 系统

### GET /status

健康检查（根路径，不在 `/api/v1/` 下）。

```json
{ "status": "ok" }
```

### POST /api/v1/stop

终止所有运行中的脚本进程，然后触发服务器优雅关闭。

```json
{ "message": "server is stopping" }
```

### GET /api/v1/getConfig

```json
{ "message": "success", "data": { "fontSize": 14 } }
```

### POST /api/v1/updateConfig

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| fontSize | fontSize | fontSize | int | 是 |

- 成功：`{ "message": "success", "data": { "fontSize": 14 } }`
- 绑定失败：`400 { "message": "invalid arguments", "data": "..." }`
- 写入失败：`500 { "message": "update config failed", "data": "..." }`

## 脚本

### GET /api/v1/scripts/

返回全部脚本，按名称排序。

`data` 为 `Script[]`：

```json
{
  "id": "uuid",
  "name": "build",
  "workDir": "C:\\repo",
  "command": ["npm", "run", "build"],
  "environments": ["env-id-1"]
}
```

- 失败：`500 { "message": "list scripts failed", "data": "..." }`

### POST /api/v1/scripts/add

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| name | name | name | string | 是 |
| 工作目录 | workdir | workDir | string | 否 |
| command | command | command | string[] | 是 |
| environmentsid | environmentsid | environmentsid | string[] | 否（要应用的环境 ID 列表） |

- 成功：`{ "message": "success", "data": 1 }`（`data` 为受影响行数）
- 参数缺失（`name` 或 `command` 为空）/ 绑定失败：`400 { "message": "invalid arguments", ... }`
- 失败：`500 { "message": "add script failed", "data": "..." }`

### PUT /api/v1/scripts/update

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| id | id | id | string | 是 |
| name | name | name | string | 是 |
| 工作目录 | workDir | workDir | string | 否 |
| command | command | command | string[] | 是 |
| environmentsid | environmentsid | environmentsid | string[] | 否 |

- 成功：`{ "message": "success", "data": { ...请求体回显 } }`
- 脚本不存在：`404 { "message": "script not found" }`
- 失败：`500 { "message": "update script failed", "data": "..." }`

### DELETE /api/v1/scripts/delete

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| id | id | id | string | 是 |

- 成功：`{ "message": "success", "data": { "id": "..." } }`
- id 为空 / 绑定失败：`400 { "message": "invalid arguments" }`
- 脚本有运行中的执行记录：`400 { "message": "script is running", "data": "script is running" }`
- 脚本不存在：`404 { "message": "script not found" }`
- 失败：`500 { "message": "delete script failed", "data": "..." }`

## 环境

### GET /api/v1/environments/

返回全部环境。

`data` 为 `Environment[]`：

```json
{
  "id": "uuid",
  "name": "java 21",
  "paths": ["C:\\jdk-21\\bin"],
  "env": [{ "key": "JAVA_HOME", "value": "C:\\jdk-21" }]
}
```

- 失败：`500 { "message": "list environments failed", "data": "..." }`

### POST /api/v1/environments/add

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| name | name | name | string | 是 |
| paths | paths | paths | string[] | 否 |
| env | env | env | `{key,value}[]` | 否 |

- 成功：`{ "message": "success", "data": 1 }`（受影响行数）
- `name` 为空 / 绑定失败：`400 { "message": "invalid arguments" }`
- 失败：`500 { "message": "add environment failed", "data": "..." }`

### PUT /api/v1/environments/update

字段同 add，另加必填的 `id`。

- 成功：`{ "message": "success", "data": 1 }`
- 环境不存在：`404 { "message": "environment not found" }`
- 失败：`500 { "message": "update environment failed", "data": "..." }`

### DELETE /api/v1/environments/delete

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| id | id | id | string | 是 |

- 成功：`{ "message": "success", "data": 1 }`
- id 为空 / 绑定失败：`400 { "message": "invalid arguments" }`
- 环境不存在：`404 { "message": "environment not found" }`
- 失败：`500 { "message": "delete environment failed" }`

## 执行

### GET /api/v1/execution/

返回全部执行记录（含已结束的，记录会保留）。

`data` 为 `ExecutionStatusInfo[]`：

```json
{
  "executionId": "uuid",
  "scriptId": "uuid",
  "name": "build",
  "startedAt": "2026-09-08T12:00:00+08:00",
  "command": ["npm", "run", "build"],
  "environments": ["JAVA_HOME=C:\\jdk-21", "Path=C:\\jdk-21\\bin;..."],
  "status": "running",
  "exitCode": -1,
  "error": ""
}
```

- `status` 取值：`prepare`（创建中）| `running` | `finished`（正常退出）| `failed`（异常/非零退出）| `killed`
- `exitCode`：运行中/被杀死时为 `-1`
- `environments`：本次运行实际应用的键值对（`K=V` 形式，已合并展开）
- 失败：`500 { "message": "list executions failed", "data": "..." }`

### POST /api/v1/execution/execute

异步启动脚本，在独立终端窗口中运行，进程独立于连接。

| 字段 | form | json | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- | --- |
| id | id | id | string | 是 | 脚本 ID |
| command | command | command | string[] | 否 | 覆盖脚本存储的命令；不传或为空时用脚本存储值 |
| environmentsid | environmentsid | environmentsid | string[] | 否 | 覆盖脚本存储的环境 ID；不传或为空时用脚本存储值 |

环境合并规则：按 ID 列表顺序应用环境组，靠后的同名变量覆盖靠前的；`paths` 去重后按序拼接到系统 `PATH` 前。

- 成功：`HTTP 200`（当前实现不写响应体，body 为空）
- id 为空 / 绑定失败：`400 { "message": "invalid arguments" }`
- 失败：`500 { "message": "start execution failed", "data": "..." }`，`data` 可能是 `script not found`、`get script fail`、`environment not found`、`get environment fail`、`execute script error`

### POST /api/v1/execution/kill

终止目标执行记录对应的进程树（`taskkill /T /F`），记录本身保留在列表中。

| 字段 | form | json | 类型 | 必填 |
| --- | --- | --- | --- | --- |
| id | id | id | string | 是（execution ID） |

- 成功：`{ "message": "success", "data": "execution deleted" }`
- id 为空 / 绑定失败：`400 { "message": "invalid arguments" }`
- 目标不在运行中（已结束/已杀死）：`400 { "message": "execution not running" }`
- 执行不存在：`500 { "message": "kill execution failed", "data": "execution not found" }`
- 其他失败：`500 { "message": "kill execution failed", "data": "..." }`

## 事件

注册事件后，事件触发时由后端按请求中携带的脚本配置异步执行脚本（同 `/execution/execute` 语义：`command`/`environmentsid` 不传时回退脚本存储值）。事件触发产生的执行错误只记录在服务端日志，不会通知前端。

> **绑定注意**：这两个请求的模型内嵌了脚本执行字段。使用 **JSON** 请求时，脚本执行字段必须嵌在 `execute_script` 对象内；使用 **form** 请求时字段平铺即可。

### GET /api/v1/event/

返回当前已注册的全部事件（文件变更事件 + 时间事件）。

`data` 为 `EventInfoResponse[]`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| eventId | string | 文件变更事件为**监听路径**（注册表按路径为键）；时间事件为生成的 UUID |
| type | string | `fileChangeEvent` \| `timeEvent` |
| event | object | 事件详情，结构因类型而异（见下） |

`event` 字段结构（注意键名为大写开头，直接序列化自内部结构体）：

文件变更事件（`type: "fileChangeEvent"`）：

```json
{ "ID": "uuid", "Path": "C:\\watch\\dir" }
```

时间事件（`type: "timeEvent"`）：

```json
{ "ID": "uuid", "Interval": 60, "Repeat": false }
```

其中 `Interval` 为秒数（与注册请求的 `interval` 一致），`Repeat` 为是否循环触发。

响应示例：

```json
{
  "message": "success",
  "data": [
    { "eventId": "C:\\watch\\dir", "type": "fileChangeEvent", "event": { "ID": "uuid", "Path": "C:\\watch\\dir" } },
    { "eventId": "uuid", "type": "timeEvent", "event": { "ID": "uuid", "Interval": 60, "Repeat": false } }
  ]
}
```

- 失败：无（总是返回 200；无事件时 `data` 为 `[]`）

### POST /api/v1/event/registerFileChangeEvent

监听文件/目录变更，变更时执行脚本。

JSON 请求体：

```json
{
  "path": "C:\\watch\\dir",
  "execute_script": {
    "id": "script-uuid",
    "command": [],
    "environmentsid": []
  }
}
```

form 请求体：`path`、`id`、`command`、`environmentsid` 平铺。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| path | string | 是 | 监听的文件/目录路径 |
| execute_script.id | string | 是 | 事件触发时要执行的脚本 ID |
| execute_script.command | string[] | 否 | 覆盖脚本存储命令 |
| execute_script.environmentsid | string[] | 否 | 覆盖脚本存储环境 ID |

- 成功：`HTTP 200`（当前实现不写响应体，body 为空）
- 绑定失败：`400 { "message": "invalid arguments", "data": "..." }`
- 路径无效（watcher 创建/添加失败）时**静默失败**：仍返回 200 空 body，但注册表条目会被回滚，事件永远不会触发，且不会向客户端报告任何错误
- 同一路径重复注册时不会新建事件，只是向已有事件追加订阅者（同一文件变更会触发多次执行）

### POST /api/v1/event/registerTimeEvent

在 `interval` 秒后执行脚本；`repeat: true` 时之后每隔 `interval` 秒重复执行。

JSON 请求体：

```json
{
  "interval": 60,
  "repeat": false,
  "execute_script": {
    "id": "script-uuid",
    "command": [],
    "environmentsid": []
  }
}
```

form 请求体：`interval`、`repeat`、`id`、`command`、`environmentsid` 平铺。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| interval | int | 是 | 间隔秒数 |
| repeat | bool | 否 | `true` 时每隔 `interval` 秒循环执行；`false`/不传时仅触发一次 |
| execute_script.id | string | 是 | 事件触发时要执行的脚本 ID |
| execute_script.command | string[] | 否 | 覆盖脚本存储命令 |
| execute_script.environmentsid | string[] | 否 | 覆盖脚本存储环境 ID |

- 成功：`HTTP 200`（当前实现不写响应体，body 为空）
- 绑定失败：`400 { "message": "invalid arguments", "data": "..." }`
