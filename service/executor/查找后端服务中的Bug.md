# 查找后端服务中的Bug

> **Workspace:** c:\develop\project\public\VeloScriptsManager

---

**User**

查找后端`c:\develop\project\public\VeloScriptsManager/service`中存在的bugs

**Assistant**

我将系统性排查后端服务代码中的逻辑缺陷。先读取所有核心文件。

继续读取服务层文件。

继续读取存储层、路由层和其余文件。

继续读取路由、入口和其余文件。

读取最后一批文件。

最后读取测试文件和 API 文档以核对预期行为。

已完成全部 33 个后端源文件的通读与分析（executor / services / storage / server / routers / events / configs / models / utils / ierrors / test）。以下是二次校验后的结构化排查报告。

# 后端逻辑 Bug 排查报告

## 概览

| # | 严重度 | 位置 | 问题摘要 |
|---|--------|------|----------|
| 1 | 高 | [execution.go](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go) | Kill 与自然退出的状态竞态，killed 可被记录为 failed |
| 2 | 高 | [execution.go](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go) | taskkill 失败时执行记录卡死，永远无法再终止 |
| 3 | 高 | [main.go](file:///c:/develop/project/public/VeloScriptsManager/service/server/main.go) | HTTP 服务启动失败后进程永久挂死 |
| 4 | 高 | [file_change_event.go](file:///c:/develop/project/public/VeloScriptsManager/service/events/file_change_event.go) | watcher.Add 失败时 registry 残留死事件，订阅永久失效 |
| 5 | 中 | [environment_router.go](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/environment_router.go) | DeleteEnvironment 删除不存在的 id 返回 200 |
| 6 | 中 | [execution_service.go](file:///c:/develop/project/public/VeloScriptsManager/service/services/execution_service.go) | KillExecution 对已结束执行返回 500，语义错误 |
| 7 | 中 | [scripts_router.go](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/scripts_router.go) | AddScript 不校验 command 为空 |
| 8 | 低 | [execution.go](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go) | finish() 非 ExitError 分支 status 为空串、exitCode 为 0 |
| 9 | 低 | [event_router.go](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/event_router.go) | RegisterFileChangeEvent 无响应体、path 不校验 |
| 10 | 低 | [caller_service.go](file:///c:/develop/project/public/VeloScriptsManager/service/services/caller_service.go) | 用户 env 中的 Path 值被 paths 静默丢弃，条件判断无效 |
| 11 | 低 | [execution_service.go](file:///c:/develop/project/public/VeloScriptsManager/service/services/execution_service.go) | Start 错误细节被吞掉 |
| 12 | 低 | [script_service.go](file:///c:/develop/project/public/VeloScriptsManager/service/services/script_service.go) | DeleteScript 存在 TOCTOU 竞态 |
| 13 | 低 | [time_event.go](file:///c:/develop/project/public/VeloScriptsManager/service/events/time_event.go) | 定时事件触发后 registry 泄漏 |
| 14 | 低 | [executor_test.go](file:///c:/develop/project/public/VeloScriptsManager/service/test/executor_test.go) | 测试引用已删除的 API，无法编译 |

---

## 高严重度

### Bug 1：Kill 与自然退出的状态竞态，killed 可能被记录为 failed

- **位置**：[execution.go#L97-L110](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go#L97-L110)、[L112-L129](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go#L112-L129)
- **表现**：`doFinish` 用 `sync.Once` 实现"先到先得"，没有 killed 状态优先级。时序：`Kill()` 持锁执行 `taskkill`（L118）→ 目标进程死亡 → `waitForExit` 中的 `finish(err)` 先在锁上排队 → `Kill` 解锁后两个 `doFinish` 竞争锁。若 `finish` 抢到（GOMAXPROCS=1、负载高、被抢占时会发生），状态被写成 `failed`。
- **影响**：用户手动终止的执行显示为红色 failed 而非橙色 killed，且 error 字段带上误导性的退出错误信息；违背"killed 状态不可被覆盖"的既定约束（当前 `sync.Once` 只保证"首个写入生效"，无法保证 killed 获胜）。
- **修复建议**：增加 `killed bool` 标志，在 `doFinish` 中优先：

```go
type Execution struct {
	...
	killed bool
}

func (execution *Execution) Kill() error {
	execution.mu.Lock()
	if execution.cmd == nil || execution.cmd.Process == nil {
		execution.mu.Unlock()
		return errors.New("process not started")
	}
	execution.killed = true
	err := exec.Command("taskkill", "/PID", strconv.Itoa(execution.cmd.Process.Pid), "/T", "/F").Run()
	execution.mu.Unlock()
	...
}

func (execution *Execution) doFinish(exitCode int, status string, err error) {
	execution.mu.Lock()
	defer execution.mu.Unlock()
	execution.doFinishOnce.Do(func() {
		if execution.killed && status != "killed" {
			status = "killed"
			exitCode = -1
		}
		...
	})
}
```

由于 `killed = true` 在 `taskkill` 之前置位，凡因 kill 导致的退出，无论哪个 `doFinish` 先到，结果均为 killed。

### Bug 2：taskkill 失败时执行记录进入错误终态，进程无法再终止

- **位置**：[execution.go#L121-L124](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go#L121-L124)
- **表现**：`taskkill` 返回错误（权限被拒等，进程实际还在运行）时，代码调用 `doFinish(-1, "failed", err)`，消耗掉 `Once` 并把 `cmd` 置为 nil。
- **影响**：① 进程仍在运行但状态显示 failed，前端终止按钮被禁用；② 后续再调 `Kill()` 因 `cmd == nil` 直接返回 "process not started"，该进程永久失去管理能力，只能手动结束；③ 附带问题：持锁执行 `taskkill.Run()`（L113-L119）期间，`GetStatus`/`List()` 等全部被阻塞数十毫秒；④ kill 与自然退出赛跑时（taskkill 报 "not found"），可能把自然成功结束记录成 failed。
- **修复建议**：taskkill 失败时不落终态、保持 running 允许重试，同时让已自然结束的 kill 幂等：

```go
if err != nil {
	return err
}
execution.doFinish(-1, "killed", nil)
return nil
```

### Bug 3：HTTP 服务启动失败后主 goroutine 永久阻塞

- **位置**：[main.go#L79-L88](file:///c:/develop/project/public/VeloScriptsManager/service/server/main.go#L79-L88)
- **表现**：`ListenAndServe` 失败（端口被占用等）只打日志，主 goroutine 仍阻塞在 `<-serverController.GetShutdownSignalChan()`。
- **影响**：进程变成僵尸——数据库已打开但不响应任何请求，`/stop` 接口也不存在（HTTP 没起来），无法优雅退出，只能强杀。Electron 侧虽做了端口探测，但 dev 模式、手动运行或探测竞态时仍会触发。
- **修复建议**：

```go
serverErr := make(chan error, 1)
go func() {
	serverErr <- server.ListenAndServe()
}()

select {
case err := <-serverErr:
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err.Error())
	}
case <-serverController.GetShutdownSignalChan():
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Println(err.Error())
	}
}

if errDbClose := db.Close(); errDbClose != nil {
	log.Println("close db: " + errDbClose.Error())
}
```

### Bug 4：watcher.Add 失败时 registry 残留死事件，订阅永久失效

- **位置**：[file_change_event.go#L65-L68](file:///c:/develop/project/public/VeloScriptsManager/service/events/file_change_event.go#L65-L68)
- **表现**：`fsnotify.NewWatcher()` 失败的分支（L57-L62）正确清理了 registry，但 `watcher.Add(path)` 失败的分支只关闭 watcher，未删除 L48 已写入的 `registry[path]`。
- **影响**：注册一个不存在的路径后，registry 中留下永远没有 watcher 的事件；之后所有对同一 path 的 `RegisterFileChangeEvent` 都走"已存在"分支追加订阅者，但这些回调永远不会被触发，调用方完全无感知（接口也不报错）；事件对象同时构成内存泄漏。
- **修复建议**：

```go
if err := watcher.Add(path); err != nil {
	watcher.Close()
	registryMu.Lock()
	delete(registry, path)
	registryMu.Unlock()
	return
}
```

---

## 中严重度

### Bug 5：DeleteEnvironment 删除不存在的 id 返回 200

- **位置**：[environment_router.go#L130-L142](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/environment_router.go#L130-L142)
- **表现**：handler 未检查 `RowsAffected == 0`（DeleteScript、UpdateEnvironment 均有 404 处理，[API.md#L105](file:///c:/develop/project/public/VeloScriptsManager/service/docs/API.md#L105) 也承诺 404）。
- **影响**：删除已被删除/不存在的环境返回 200 success，前端无法感知失败，与同类接口行为不一致。
- **修复建议**：在 `apiErr` 判断之后、返回 200 之前补上：

```go
if count == 0 {
	c.JSON(404, gin.H{
		"message": "environment not found",
	})
	return
}
```

### Bug 6：KillExecution 对已结束执行返回 500，语义错误

- **位置**：[execution_service.go#L68-L80](file:///c:/develop/project/public/VeloScriptsManager/service/services/execution_service.go#L68-L80)、[execution.go#L113-L117](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go#L113-L117)
- **表现**：对已结束（finished/failed）的执行调用 kill，`Kill()` 因 `cmd == nil` 返回 "process not started"，一路映射为 HTTP 500。另外返回文案 `"execution deleted"` 与实际行为（记录保留）不符。
- **影响**：API 语义混乱；kill 一个恰好自然退出的执行会返回 500，让用户误以为操作失败；API 层无法区分"不存在"（应为 404）与"内部错误"。
- **修复建议**：在 `KillExecution` 中先检查 `GetStatus()`，已处于终态直接返回成功或专用错误（映射 404/409）；`Kill` 对已终态执行返回 nil 实现幂等；响应文案改为与实际行为一致的描述。

### Bug 7：AddScript 不校验 command 为空

- **位置**：[scripts_router.go#L64-L70](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/scripts_router.go#L64-L70)
- **表现**：只校验 Name 非空，`command` 为空/null 也能入库（API.md 标注 command 必填）。
- **影响**：可以创建出永远无法执行的脚本，运行时才报错，且报错被 Bug 11 吞成泛化信息，用户无从得知原因。
- **修复建议**：与 Name 校验并列增加 `len(req.Command) == 0` 的 400 分支。

---

## 低严重度

### Bug 8：finish() 非 ExitError 分支产生空 status 和 exitCode=0

- **位置**：[execution.go#L138-L144](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution.go#L138-L144)
- **表现**：`err != nil` 但不是 `*exec.ExitError` 时，`status` 保持零值 `""`、`exitCode` 保持 0。分支覆盖不全。
- **影响**：该执行在列表中显示空状态，前端状态配色/排序/终止按钮判断全部失效；exitCode=0 还伪装成成功。
- **修复建议**：else 分支兜底 `status = "failed"`，exitCode 取 -1（或保留 exitErr.ExitCode() 的分支并给 default 赋值）。

### Bug 9：RegisterFileChangeEvent 无响应体且不校验 path

- **位置**：[event_router.go#L26-L45](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/event_router.go#L26-L45)
- **表现**：handler 处理完不写任何响应（200 空 body）；`req.Path` 为空也不拦截。
- **影响**：前端 api client 校验 `data` 字段缺失会直接抛错；空 path 走到 `watcher.Add("")` 静默失败，叠加 Bug 4 注册形同虚设且无任何反馈。
- **修复建议**：补 `req.Path == ""` 的 400 校验；成功后返回 `c.JSON(200, gin.H{"message": "success", "data": ...})` 保持信封一致。

### Bug 10：resolveEnvironmentVars 中 `key != "Path"` 条件无效，用户 Path 值被静默丢弃

- **位置**：[caller_service.go#L105-L117](file:///c:/develop/project/public/VeloScriptsManager/service/services/caller_service.go#L105-L117)
- **表现**：循环删除非 "Path" 拼写的 path 变量，但随后 `values["Path"] = strings.Join(merged, ";")` 无条件覆盖，精确等于 "Path" 的用户值同样被丢弃——`key != "Path"` 判断形同虚设。而 paths 为空时（不进入该 if），env 中的 Path 又会被保留并在 mergeEnviron 中前置拼接。
- **影响**：两种场景行为不一致；用户在 env 中配置的 Path 值在配置了 paths 时被静默丢失，且代码意图（保留 "Path"）与实际行为不符。
- **修复建议**：明确语义后二选一：① 把用户 env 的 Path 值并入合并（如 prepend 到 merged）；② paths 为空时同样忽略 env 中的 Path，保证两个分支一致，同时删除无效的 `key != "Path"` 条件。

### Bug 11：Start 的错误细节被统一吞掉

- **位置**：[execution_service.go#L47-L51](file:///c:/develop/project/public/VeloScriptsManager/service/services/execution_service.go#L47-L51)
- **表现**：`Start` 的具体错误（"empty command"、找不到可执行文件、工作目录无效等）全部替换为 `ierrors.ExecuteScriptError`。
- **影响**：用户启动失败时只能看到 "execute script error"，无法定位是命令为空还是路径错误，排障成本高。
- **修复建议**：`return nil, fmt.Errorf("%w: %v", ierrors.ExecuteScriptError, startErr)` 保留细节。

### Bug 12：DeleteScript 存在 TOCTOU 竞态

- **位置**：[script_service.go#L91-99](file:///c:/develop/project/public/VeloScriptsManager/service/services/script_service.go#L91-L99)
- **表现**：检查"无 running 执行"与执行 DB 删除之间没有原子性，另一个并发请求可在此间隙启动该脚本。
- **影响**：脚本被删除的同时仍被执行，产生引用已删脚本的运行记录（影响较轻，仅数据一致性）。
- **修复建议**：接受该窗口并在启动侧校验兜底（`MakeAndStartExecution` 已查脚本存在性，实际已闭环），或引入脚本级启动锁收紧窗口。

### Bug 13：time_event 触发后 registry 条目泄漏

- **位置**：[time_event.go#L39-L47](file:///c:/develop/project/public/VeloScriptsManager/service/events/time_event.go#L39-L47)
- **表现**：定时器触发后未从 `timeEventRegistry` 删除事件；`registerTimeEvent` 为未导出函数且当前无调用方（死代码）。
- **影响**：若未来启用，每次注册泄漏一个条目；当前无实际运行影响。
- **修复建议**：goroutine 末尾加 `timeEventRegistryLock.Lock(); delete(timeEventRegistry, timeEvent.ID); timeEventRegistryLock.Unlock()`，或直接移除死代码。

### Bug 14：测试引用已删除的 API，测试包无法编译

- **位置**：[executor_test.go#L14](file:///c:/develop/project/public/VeloScriptsManager/service/test/executor_test.go#L14)、L32、L62
- **表现**：测试调用 `executor.Start(...)`、`process.Done()/ExitCode()/Err()`，这些 API 已随 [process.go](file:///c:/develop/project/public/VeloScriptsManager/service/executor/process.go) 全量注释而删除。
- **影响**：`go test ./...` 在该包直接编译失败，`TestStartEmptyCommand`/`TestStartProcessExitCode`/`TestStartProcessEnv` 全部失效。
- **修复建议**：基于新的 `Execution` API 重写（NewExecution + Start + GetExitCode/GetStatus），或删除过时用例。

---

## 二次校验说明

**已确认生效的历史修复**（未重复报告）：registry 的 `registryMu` 保护、watcher goroutine 中订阅者拷贝、check-then-act 原子化、`KillExecution` 不再删除记录、`doFinish` 的 nil err 检查。

**核实后排除的疑似问题**（避免误报）：
- `startExecution` 成功无响应体——前后端既定约定；
- 执行记录内存保留（不持久化、不删除）——设计取舍，仅提示 [execution_manager.go](file:///c:/develop/project/public/VeloScriptsManager/service/executor/execution_manager.go#L20-L24) 长驻运行会无限增长，属设计层面而非缺陷；
- CORS 放行 `null` origin——为打包后 `file://` 前端所需；
- `mergeEnviron` 的 PATH 前置拼接、`environmentsid` 小写 JSON tag——均为项目约定；
- `Start` 中无锁写 `cmd/status`——发生在加入 ExecutionManager 之前，无并发读者；
- [scripts_router.go#L175-L218](file:///c:/develop/project/public/VeloScriptsManager/service/server/routers/scripts_router.go#L175-L218) 的 `ExecuteScript` 为已注册路由之外的死代码，无运行时影响（可在后续清理）。

建议优先处理 Bug 1-4（涉及进程生命周期与状态正确性），Bug 5/6/7 可一并小改动收敛。需要我直接修复其中某几项时告诉我编号即可（遵循不加注释、不动原注释的约束）。

---

