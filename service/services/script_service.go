package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/emirpasic/gods/sets/linkedhashset"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

type ScriptService struct {
	scriptRepo         *storage.ScriptRepo
	executions         *executor.ExecutionManager
	environmentService *EnvironmentService
}

func NewScriptService(scriptRepo *storage.ScriptRepo, manager *executor.ExecutionManager, environmentService *EnvironmentService) *ScriptService {
	return &ScriptService{
		scriptRepo:         scriptRepo,
		executions:         manager,
		environmentService: environmentService,
	}
}

func (service *ScriptService) ListScripts() (*models.Result, *models.ApiError) {
	list, err := service.scriptRepo.List()
	if err != nil {
		return nil, models.NewApiError(500, "db list error", err)
	}
	return models.NewResult(list), nil
}

func (service *ScriptService) AddScript(req *models.AddScriptRequest) (*models.Result, *models.ApiError) {

	script := storage.Script{
		ID:           utils.GenerateScriptId(),
		Name:         req.Name,
		WorkDir:      req.WorkDir,
		Command:      req.Command,
		Environments: req.EnvironmentsId,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return nil, models.NewApiError(500, "upsert script fail", err.Error())
	}

	return models.NewResultWithMessage("script added", nil), nil
}

func (service *ScriptService) UpdateScript(req *models.UpdateScriptRequest) (*models.Result, *models.ApiError) {
	if req.Id == "" {
		return nil, models.NewApiError(400, "invalid arguments", "id cannot be empty")
	}

	script := storage.Script{
		ID:           req.Id,
		Name:         req.Name,
		WorkDir:      req.WorkDir,
		Command:      req.Command,
		Environments: req.EnvironmentsId,
	}

	if err := service.scriptRepo.Update(script); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "script not found", req.Id)
		}
		return nil, models.NewApiError(500, "update script fail", err.Error())
	}

	return models.NewResultWithMessage("script updated", nil), nil
}

func (service *ScriptService) DeleteScript(id string) (*models.Result, *models.ApiError) {
	if execution, ok := service.executions.Get(id); ok {
		if execution.Status() == "running" {
			return nil, models.NewApiError(406, "script is running", execution)
		}
	}

	if err := service.scriptRepo.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "script not found", id)
		}
		return nil, models.NewApiError(500, "delete script fail", err.Error())
	}
	return models.NewResultWithMessage("script deleted", nil), nil
}

// StartExecution 加载脚本并执行。他会提供execution句柄，用作attach。
// 执行时允许前端覆盖脚本存储的 Params / Environments：请求体里传了就使用传递的值，
// 没传则回退到脚本自身存储的配置。
func (service *ScriptService) StartExecution(req models.ExecuteScriptRequest) (*models.Execution, *models.ApiError) {
	script, err := service.scriptRepo.Get(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewApiError(404, "script not found", err.Error())
		}
		return nil, models.NewApiError(http.StatusNotFound, "get script fail", err.Error())
	}

	// 优先使用前端传递的覆盖参数，否则回退到脚本存储值。
	command := script.Command
	if req.Command != nil {
		command = req.Command
	}
	envIDs := script.Environments
	if req.EnvironmentsId != nil {
		envIDs = req.EnvironmentsId
	}

	env, apiErr := service.resolveEnvironmentVars(envIDs)
	if apiErr != nil {
		return nil, apiErr
	}

	//launch background script
	process, err := executor.Start(
		context.Background(),
		script.Name,
		command,
		script.WorkDir,
		env,
	)
	if err != nil {
		return nil, models.NewApiError(500, "start script fail", err.Error())
	}

	execution := models.NewExecution(utils.GenerateExecutionId(), script.ID, script.Name, command, envIDs, process)
	service.executions.Add(execution)

	go func() {
		<-process.Done()
		execution.Finish(process.ExitCode(), process.Err())
	}()

	return execution, nil
}

// GetExecution returns the tracked execution by id, or a 404 when unknown.
func (service *ScriptService) GetExecution(id string) (*models.Execution, *models.ApiError) {
	execution, ok := service.executions.Get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", nil)
	}
	return execution, nil
}

// ListExecutions returns an id/status snapshot of every tracked execution,
// ordered by start time. Records persist after the process exits until they
// are removed explicitly via DeleteExecution.
func (service *ScriptService) ListExecutions() (*models.Result, *models.ApiError) {
	executions := service.executions.List()

	list := make([]models.ExecutionStatusInfo, 0, len(executions))
	for _, e := range executions {
		list = append(list, models.ExecutionStatusInfo{
			ExecutionId:  e.ID,
			ScriptId:     e.ScriptID,
			Name:         e.Name,
			StartedAt:    e.StartedAt,
			Command:      e.Command,
			Environments: e.Environments,
			Status:       e.Status(),
			ExitCode:     e.ExitCode(),
			Error:        e.Error(),
		})
	}

	return models.NewResult(list), nil
}

// DeleteExecution removes a tracked execution record by id. A still-running
// process is killed first so its handle is never lost mid-flight.
func (service *ScriptService) DeleteExecution(id string) (*models.Result, *models.ApiError) {
	execution, ok := service.executions.Get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", id)
	}

	if p := execution.Process(); p != nil {
		_ = p.Kill()
	}

	service.executions.Remove(id)
	return models.NewResultWithMessage("execution deleted", nil), nil
}

// resolveEnvironmentVars 展平所有的链式依赖，列表后续的 env 可以覆盖前面的同名变量。
func (service *ScriptService) resolveEnvironmentVars(ids []string) ([]string, *models.ApiError) {
	if len(ids) == 0 {
		return nil, nil
	}

	values := make(map[string]string)
	paths := linkedhashset.New()

	for _, id := range ids {
		environment, err := service.environmentService.getEnvironment(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, models.NewApiError(404, "environment not found", id)
			}
			return nil, models.NewApiError(500, "get environment fail", err.Error())
		}

		for _, p := range environment.Paths {
			if p == "" {
				continue
			}
			paths.Add(p)
		}

		for _, item := range environment.Env {
			values[item.Key] = item.Value
		}
	}

	if paths.Size() > 0 {
		for key := range values {
			if strings.EqualFold(key, "path") && key != "Path" {
				delete(values, key)
			}
		}
		rawPaths := paths.Values()
		merged := make([]string, 0, len(rawPaths))
		for _, p := range rawPaths {
			merged = append(merged, p.(string))
		}
		values["Path"] = strings.Join(merged, ";")
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+values[key])
	}
	return env, nil
}
