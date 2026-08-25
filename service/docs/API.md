
## 注意！！！
    这个api文档由ai生成。我暂时没精力手写。仅供参考。以实际代码为准

## 通用约定

### 响应信封格式

所有接口统一返回如下结构（HTTP 状态码与 `code` 字段一致）：

```json
// 成功
{ "code": 200, "message": "ok", "data": ... }

// 失败
{ "code": 400, "message": "invalid arguments", "data": ... }
```

### GET /status

健康检查。

**响应**

```json
{ "status": "ok" }
```

> 注意：此接口在 `/api/v1/` 之外，直接位于根路径。

### POST /api/v1/stop

优雅停止服务器（触发 shutdown）。

**响应**

```json
{ "code": 200, "message": "server is stopping" }
```

---

## 脚本管理

### GET /api/v1/getStoredScripts

获取所有已存储的脚本列表。


```json
[
  {
    "id": "scr-xxxx",
    "name": "build",
    "workDir": "C:\\repo",
    "runner": "npm",
    "params": ["run", "build"],
    "environments": ["id"]
  }
]
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 脚本 ID（服务端生成） |
| name | string | 脚本名称 |
| workDir | string | 运行时工作目录 |
| runner | string | 可执行程序（runner） |
| params | string[] | 传给 runner 的参数 |
| environments | string[] | 运行前应用的 environment ID 列表 |

### POST /api/v1/addScript

新增脚本（ID 由服务端生成后 Upsert）。

**请求参数**

| 字段 | 类型 | 必填 | 说明                         |
| --- | --- | --- |------------------------------|
| name | string | 是 | 脚本名称                     |
| workdir | string | 是 | 工作目录                     |
| runner | string | 是 | 可执行程序                   |
| params | string[] | 否 | 参数列表                     |
| environments | string[] | 否 | 要应用的 environment ID 列表 |

**请求示例**

```json
{
  "name": "build",
  "workdir": "C:\\repo",
  "runner": "npm",
  "params": ["run", "build"],
  "environments": ["env-xxxx"]
}
```

**响应**

```json
{ "code": 200, "message": "script added", "data": null }
```

### POST /api/v1/deleteScript

按 ID 删除脚本。

**请求参数**

| 字段 | 类型 | 必填 |
| --- | --- | --- |
| id | string | 是 |

**响应**

```json
{ "code": 200, "message": "script deleted", "data": null }
```

---

## 环境管理

### GET /api/v1/getEnvironments

获取所有环境列表。

**响应 `data` 字段**：`Environment` 数组

```json
[
  {
    "id": "env-xxxx",
    "name": "node 20",
    "type": "runtime",
    "path": "C:\\nodejs",
    "env": [{ "key": "NODE_ENV", "value": "production" }],
    "children": ["env-yyyy"]
  }
]
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 环境 ID（服务端生成） |
| name | string | 环境名称 |
| type | string | 环境类型 |
| path | string | 关联路径 |
| env | EnvVar[] | 该环境贡献的键值对 |
| children | string[] | 继承的其他环境 ID（先应用 children，后应用自身 env 覆盖） |

### POST /api/v1/addEnvironment

新增环境。

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| name | string | 是 | 环境名称 |
| type | string | 是 | 环境类型 |
| path | string | 是 | 关联路径 |
| env | `{key, value}[]` | 否 | 环境变量键值对 |
| children | string[] | 否 | 继承的 environment ID 列表 |

**响应**

```json
{ "code": 200, "message": "environment added", "data": null }
```

### POST /api/v1/deleteEnvironment

按 ID 删除环境。

**请求参数**：同 deleteScript（`id`）。

**响应**

```json
{ "code": 200, "message": "environment deleted", "data": null }
```

---



### POST /api/v1/executeScript

异步启动脚本，立即返回 execution 信息。进程独立于 HTTP 连接运行，即使客户端断开也会运行至结束。

启动前会解析脚本引用的 environments（递归展开 children），合并为环境变量传入进程：
- 优先级为「后写覆盖」：children（基础）先应用，环境自身的 env 覆盖之；脚本 environments 列表中靠后的 ID 覆盖靠前的。
- children 引用成环时返回 500 `environment cycle detected`。

**请求参数**

| 字段 | 类型 | 必填 |
| --- | --- | --- |
| id | string | 是（脚本 ID） |

**响应 `data` 字段**

```json
{
  "executionId": "uuid",
  "scriptId": "uuid",
  "name": "build"
}
```

**错误**

| code | message |
| --- | --- |
| 404 | script not found |
| 404 | environment not found（`data` 为对应 ID） |
| 500 | start script fail / environment cycle detected |

### GET /api/v1/execute/attach?executionId=... （WebSocket）

将连接升级为 WebSocket，桥接到运行中执行进程的 stdio。

**Query 参数**

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| executionId | 是 | executeScript 返回的执行 ID；未知时返回 404 `execution not found` |

**服务端 → 客户端帧**

```json
{ "type": "output", "data": "<base64 输出块>" }   // 进程 stdout/stderr 分块，base64 编码（保证非 UTF-8 如 GBK 输出完整传输）
{ "type": "exit", "code": 0, "error": "" }        // 进程结束；error 非空表示启动/运行出错
```

**客户端 → 服务端帧**

```json
{ "type": "stdin", "data": "<base64 字节>" }  // 写入进程 stdin
{ "type": "close_stdin" }                     // 关闭 stdin（对交互式程序即 EOF）
{ "type": "kill" }                            // 强制终止进程
```

进程结束后服务端发送 `exit` 帧并关闭连接。
