## 通用约定

所有接口统一返回 `{code, message, data}` 信封（`code` 同时作为 HTTP 状态码）。
- 成功：`{ "code": 200, "message": "ok", "data": ... }`
- 失败：`{ "code": 400, "message": "...", "data": ... }`

请求体支持 `form` 与 `json` 两种绑定方式（字段名见各接口）。

## 系统

### GET /status
健康检查，返回 `{ "status": "ok" }`（根路径，不在 `/api/v1/` 下）。

### POST /api/v1/stop
触发服务器优雅关闭，返回 `{ "code": 200, "message": "server is stopping" }`。

## 脚本

### GET /api/v1/getStoredScripts
返回 `Script[]`。

```json
{ "id": "scr-xxx", "name": "build", "workDir": "C:\\repo",
  "command": ["npm","run","build"], "environments": ["env-id"] }
```

### POST /api/v1/addScript
| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| name | string | 是 | 脚本名称 |
| workdir | string | 是 | 工作目录 |
| command | string[] | 是 | 有序命令节点，第一个节点为程序 |
| environmentsid | string[] | 否 | 应用的环境 ID 列表 |

返回 `{ "code": 200, "message": "script added" }`。

### POST /api/v1/updateScript
同 addScript，但 `id` 必填（json 字段 `id`）；返回 `message: "script updated"`。脚本不存在返回 `404 script not found`。

### POST /api/v1/deleteScript
| 字段 | 类型 | 必填 |
| --- | --- | --- |
| id | string | 是 |

返回 `message: "script deleted"`；不存在 `404 script not found`。

### POST /api/v1/executeScript
异步启动脚本，立即返回 execution 信息（进程独立于连接运行）。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| id | string | 是 | 脚本 ID |
| command | string[] | 否 | 覆盖脚本存储命令节点 |
| environmentsid | string[] | 否 | 覆盖脚本存储环境 ID |

返回 `data`:
```json
{ "executionId": "uuid", "scriptId": "uuid", "name": "build" }
```
启动前合并 environments（列表靠后的覆盖靠前的同名变量；paths 去重后按序拼接到系统 PATH 前）。
错误：`404 script not found` / `404 environment not found` / `500 start script fail`。

### GET /api/v1/getExecutions
返回 `ExecutionStatusInfo[]`（按启动时间排序，进程结束后保留记录直到删除）：
```json
{ "executionId":"uuid","scriptId":"uuid","name":"build",
  "startedAt":"2026-08-26T12:00:00Z","status":"running|finished|failed",
  "exitCode":-1,"error":"" }
```

### POST /api/v1/deleteExecution
| 字段 | 类型 | 必填 |
| --- | --- | --- |
| id | string | 是（execution ID） |

仍在运行的进程会先强制终止；返回 `message: "execution deleted"`；不存在 `404 execution not found`。

## 环境

### GET /api/v1/getEnvironments
返回 `Environment[]`：
```json
{ "id":"env-xxx","name":"java 21",
  "paths":["C:\\jdk-21\\bin"],
  "env":[{"key":"JAVA_HOME","value":"C:\\jdk-21"}] }
```

### POST /api/v1/addEnvironment
| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| name | string | 是 | 环境名称 |
| paths | string[] | 是 | 关联路径，执行时拼接到系统 PATH 前 |
| env | `{key,value}[]` | 否 | 环境变量，按后覆盖前合并 |

返回 `message: "environment added"`。

### POST /api/v1/updateEnvironment
同 addEnvironment，但 `id` 必填（json 字段 `id`）；返回 `message: "environment updated"`。环境不存在返回 `404 environment not found`。

### POST /api/v1/deleteEnvironment
| 字段 | 类型 | 必填 |
| --- | --- | --- |
| id | string | 是 |

返回 `message: "environment deleted"`；不存在 `404 environment not found`。

## 设置

### GET /api/v1/getConfig
返回 `Config`：
```json
{ "fontSize": 14 }
```

### POST /api/v1/updateConfig
| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| fontSize | int | 是 | 字体大小 |

返回 `message: "config updated"`。
