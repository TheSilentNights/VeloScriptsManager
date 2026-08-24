package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
		ID:      utils.GenerateScriptId(),
		Name:    req.Name,
		WorkDir: req.WorkDir,
		Runner:  req.Runner,
		Params:  req.Params,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return nil, models.NewApiError(500, "upsert script fail", err.Error())
	}

	return models.NewResultWithMessage("script added", nil), nil
}

func (service *Service) DeleteScript(id string) (*models.Result, *models.ApiError) {
	if err := service.scriptRepo.Delete(id); err != nil {
		return nil, models.NewApiError(500, "delete script fail", err.Error())
	}
	return models.NewResultWithMessage("script deleted", nil), nil
}

// StartExecution loads the script identified by id and launches it
// asynchronously via the executor, using the script's own runner, params and
// work dir. It returns immediately with an Execution handle; callers can later
// attach to the process stdio through the execution id.
func (service *Service) StartExecution(id string) (*Execution, *models.ApiError) {
	script, err := service.scriptRepo.Get(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewApiError(404, "script not found", err.Error())
		}
		return nil, models.NewApiError(500, "get script fail", err.Error())
	}

	// context.Background keeps the process running to completion even if the
	// HTTP client that started it disconnects.
	process, err := executor.Start(context.Background(), script.Runner, script.Params, script.WorkDir)
	if err != nil {
		return nil, models.NewApiError(500, "start script fail", err.Error())
	}

	execution := &Execution{
		ID:        utils.GenerateExecutionId(),
		ScriptID:  script.ID,
		Name:      script.Name,
		Status:    "running",
		ExitCode:  -1,
		StartedAt: time.Now(),
		Process:   process,
	}
	service.executions.add(execution)

	go func() {
		<-process.Done()
		execution.ExitCode = process.ExitCode()
		if execErr := process.Err(); execErr != nil {
			execution.Status = "failed"
			execution.Error = execErr.Error()
		} else {
			execution.Status = "finished"
		}
	}()

	return execution, nil
}

// GetExecution returns the tracked execution by id, or a 404 when unknown.
func (service *Service) GetExecution(id string) (*Execution, *models.ApiError) {
	execution, ok := service.executions.get(id)
	if !ok {
		return nil, models.NewApiError(404, "execution not found", nil)
	}
	return execution, nil
}
