package services

import (
	"context"
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"log"
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
	scriptOptions *[]models.ScriptOption,
) ([]*executor.Execution, error) {
	if scriptOptions == nil || len(*scriptOptions) == 0 {
		return nil, ierrors.InvalidScriptOption
	}

	options := *scriptOptions

	firstFinishChan := make(chan struct{})
	first := executor.NewExecution(
		options[0].ID,
		options[0].ScriptName,
		options[0].WorkDir,
		options[0].Command,
		options[0].Env,
		firstFinishChan,
	)

	startErr := first.Start(context.Background())

	if startErr != nil {
		log.Printf("start execution failed: %v", startErr)
		return nil, ierrors.ExecuteScriptError
	}

	service.executions.Add(first)

	if len(options) > 1 {
		go service.runExecutionChain(first, firstFinishChan, options[1:])
	}

	return []*executor.Execution{first}, nil
}

func (service *ExecutionService) runExecutionChain(
	prev *executor.Execution,
	prevFinishChan <-chan struct{},
	options []models.ScriptOption,
) {
	for _, option := range options {
		<-prevFinishChan

		if prev.GetStatus() != "finished" {
			log.Printf("execution chain stopped: execution %s ended with status %s",
				prev.GetExecutionId(), prev.GetStatus())
			return
		}

		finishChan := make(chan struct{})
		execution := executor.NewExecution(
			option.ID,
			option.ScriptName,
			option.WorkDir,
			option.Command,
			option.Env,
			finishChan,
		)

		startErr := execution.Start(context.Background())
		if startErr != nil {
			log.Printf("start execution failed: %v", startErr)
			return
		}

		service.executions.Add(execution)
		prev = execution
		prevFinishChan = finishChan
	}
}

func waitForExecutionFinish(channel chan struct{}, callback func() error) error {

	<-channel

	err := callback()
	if err != nil {
		return err
	}

	return nil
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
