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

func (service *Service) ListScripts() (*models.Result, *models.ApiError) {
	list, err := service.scriptRepo.List()
	if err != nil {
		return nil, models.NewApiError(500, "db list error", err)
	}
	return models.NewResult(list), nil
}

func (service *Service) AddScript(req *models.AddScriptRequest) (*models.Result, *models.ApiError) {

	script := storage.Script{
		ID:           utils.GenerateScriptId(),
		Name:         req.Name,
		WorkDir:      req.WorkDir,
		Runner:       req.Runner,
		Params:       req.Params,
		Environments: req.Environments,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return nil, models.NewApiError(500, "upsert script fail", err.Error())
	}

	return models.NewResultWithMessage("script added", nil), nil
}

func (service *Service) DeleteScript(id string) (*models.Result, *models.ApiError) {
	if err := service.scriptRepo.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "script not found", id)
		}
		return nil, models.NewApiError(500, "delete script fail", err.Error())
	}
	return models.NewResultWithMessage("script deleted", nil), nil
}

// StartExecution 加载脚本并执行。他会提供execution句柄，用作attach
func (service *Service) StartExecution(id string) (*models.Execution, *models.ApiError) {
	script, err := service.scriptRepo.Get(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewApiError(404, "script not found", err.Error())
		}
		return nil, models.NewApiError(500, "get script fail", err.Error())
	}

	env, apiErr := service.resolveEnvironmentVars(script.Environments)
	if apiErr != nil {
		return nil, apiErr
	}

	//launch background script
	process, err := executor.Start(context.Background(), script.Runner, script.Params, script.WorkDir, env)
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
func (service *Service) GetExecution(id string) (*models.Execution, *models.ApiError) {
	execution, ok := service.executions.get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", nil)
	}
	return execution, nil
}

// ListExecutions returns an id/status snapshot of every tracked execution,
// ordered by start time. Records persist after the process exits until they
// are removed explicitly via DeleteExecution.
func (service *Service) ListExecutions() (*models.Result, *models.ApiError) {
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
func (service *Service) DeleteExecution(id string) (*models.Result, *models.ApiError) {
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
func (service *Service) resolveEnvironmentVars(ids []string) ([]string, *models.ApiError) {
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

		environment, err := service.environmentRepo.Get(id)
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
