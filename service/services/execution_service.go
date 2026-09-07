package services

import (
	"context"
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
)

var executionService *ExecutionService

type ExecutionProvider interface {
	List() []*executor.Execution
}

type ExecutionService struct {
	executions *executor.ExecutionManager
}

func NewExecutionService(executions *executor.ExecutionManager) *ExecutionService {
	executionService = &ExecutionService{
		executions: executions,
	}
	return executionService
}

func GetExecutionProvider() ExecutionProvider {
	return executionService
}

func (service *ExecutionService) MakeAndStartExecution(
	id string,
	scriptName string,
	command []string,
	workDir string,
	env []string,
) (*executor.Execution, error) {

	execution := executor.NewExecution(
		id,
		scriptName,
		command,
		workDir,
		env,
	)

	startErr := execution.Start(context.Background())

	if startErr != nil {
		return nil, ierrors.ExecuteScriptError
	}

	service.executions.Add(execution)

	return execution, nil
}

func (service *ExecutionService) GetExecution(id string) (*executor.Execution, error) {
	execution, ok := service.executions.Get(id)
	if !ok {
		return nil, ierrors.ExecutionNotFound
	}
	return execution, nil
}

// KillExecution removes a tracked execution record by id. A still-running
// process is killed first so its handle is never lost mid-flight.
func (service *ExecutionService) KillExecution(id string) (any, error) {
	execution, ok := service.executions.Get(id)
	if !ok {
		return nil, ierrors.ExecutionNotFound
	}

	killErr := execution.Kill()

	if killErr != nil {
		return nil, killErr
	}
	service.executions.Remove(id)
	return "execution deleted", nil
}

func (service *ExecutionService) List() []*executor.Execution {
	return service.executions.List()
}

func (service *ExecutionService) ListExecutions() (any, error) {
	executions := service.executions.List()

	list := make([]models.ExecutionStatusInfo, 0, len(executions))
	for _, e := range executions {
		list = append(list, models.ExecutionStatusInfo{
			ExecutionId:  e.GetExecutionId(),
			ScriptId:     e.GetScriptInfo().ScriptID,
			Name:         e.GetScriptInfo().Name,
			StartedAt:    e.GetScriptInfo().StartedAt,
			Command:      e.GetScriptInfo().Command,
			Environments: e.GetScriptInfo().EnvironmentsFlattened,
			Status:       e.GetStatus(),
			ExitCode:     e.GetExitCode(),
			Error:        e.GetError(),
		})
	}

	return list, nil
}
