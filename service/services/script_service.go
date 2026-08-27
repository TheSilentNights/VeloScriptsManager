package services

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

type ScriptService struct {
	scriptRepo         *storage.ScriptRepo
	executions         *ExecutionManager
	environmentService *EnvironmentService
}

func NewScriptService(scriptRepo *storage.ScriptRepo, manager *ExecutionManager, environmentService *EnvironmentService) *ScriptService {
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
		Runner:       req.Runner,
		Params:       req.Params,
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
		Runner:       req.Runner,
		Params:       req.Params,
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
		return nil, models.NewApiError(500, "get script fail", err.Error())
	}

	// 优先使用前端传递的覆盖参数，否则回退到脚本存储值。
	params := script.Params
	if req.Params != nil {
		params = req.Params
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
	process, err := executor.Start(context.Background(), script.Runner, params, script.WorkDir, env)
	if err != nil {
		return nil, models.NewApiError(500, "start script fail", err.Error())
	}

	execution := models.NewExecution(utils.GenerateExecutionId(), script.ID, script.Name, process)
	service.executions.add(execution)

	go func() {
		<-process.Done()
		execution.Finish(process.ExitCode(), process.Err())
	}()

	return execution, nil
}

// GetExecution returns the tracked execution by id, or a 404 when unknown.
func (service *ScriptService) GetExecution(id string) (*models.Execution, *models.ApiError) {
	execution, ok := service.executions.get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", nil)
	}
	return execution, nil
}

// ListExecutions returns an id/status snapshot of every tracked execution,
// ordered by start time. Records persist after the process exits until they
// are removed explicitly via DeleteExecution.
func (service *ScriptService) ListExecutions() (*models.Result, *models.ApiError) {
	executions := service.executions.list()

	list := make([]models.ExecutionStatusInfo, 0, len(executions))
	for _, e := range executions {
		list = append(list, models.ExecutionStatusInfo{
			ExecutionId: e.ID,
			ScriptId:    e.ScriptID,
			Name:        e.Name,
			StartedAt:   e.StartedAt,
			Status:      e.Status(),
			ExitCode:    e.ExitCode(),
			Error:       e.Error(),
		})
	}

	return models.NewResult(list), nil
}

// DeleteExecution removes a tracked execution record by id. A still-running
// process is killed first so its handle is never lost mid-flight.
func (service *ScriptService) DeleteExecution(id string) (*models.Result, *models.ApiError) {
	execution, ok := service.executions.get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", id)
	}

	if p := execution.Process(); p != nil {
		_ = p.Kill()
	}

	service.executions.remove(id)
	return models.NewResultWithMessage("execution deleted", nil), nil
}

// resolveEnvironmentVars 展平所有的链式依赖，列表后续的 env 可以覆盖前面的同名变量。
func (service *ScriptService) resolveEnvironmentVars(ids []string) ([]string, *models.ApiError) {
	if len(ids) == 0 {
		return nil, nil
	}

	values := make(map[string]string)
	applied := make(map[string]bool)
	/*
		在一个调用链重，一个envid不能同时出现两次，否则是链内循环，因而写stack。
	*/
	stack := make(map[string]bool)

	var visit func(id string, isTop bool) *models.ApiError
	visit = func(id string, isTop bool) *models.ApiError {
		if stack[id] {
			return models.NewApiError(500, "environment cycle detected", id)
		}
		stack[id] = true
		defer delete(stack, id)

		environment, err := service.environmentService.getEnvironment(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return models.NewApiError(404, "environment not found", id)
			}
			return models.NewApiError(500, "get environment fail", err.Error())
		}

		for _, child := range environment.Children {
			if apiErr := visit(child, false); apiErr != nil {
				return apiErr
			}
		}

		if isTop || !applied[id] {
			for _, item := range environment.Env {
				values[item.Key] = item.Value
			}
			applied[id] = true
		}

		return nil
	}

	for _, id := range ids {
		if apiErr := visit(id, true); apiErr != nil {
			return nil, apiErr
		}
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
