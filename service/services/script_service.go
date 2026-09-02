package services

import (
	"context"
	"database/sql"
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"sort"
	"strings"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"

	"github.com/emirpasic/gods/sets/linkedhashset"
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

func (service *ScriptService) ListScripts() (any, error) {
	list, err := service.scriptRepo.List()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (service *ScriptService) AddScript(req *models.AddScriptRequest) (int64, error) {

	script := storage.Script{
		ID:           utils.GenerateScriptId(),
		Name:         req.Name,
		WorkDir:      req.WorkDir,
		Command:      req.Command,
		Environments: req.EnvironmentsId,
	}

	count, err := service.scriptRepo.Insert(script)

	if err != nil {
		return -1, errors.New("db add error")
	}

	return count, nil
}

func (service *ScriptService) UpdateScript(req *models.UpdateScriptRequest) (int64, error) {

	script := storage.Script{
		ID:           req.Id,
		Name:         req.Name,
		WorkDir:      req.WorkDir,
		Command:      req.Command,
		Environments: req.EnvironmentsId,
	}
	count, err := service.scriptRepo.Update(script)

	if err != nil {
		return -1, ierrors.UpdateScriptDbError
	}

	return count, nil
}

func (service *ScriptService) DeleteScript(id string) (int64, error) {

	findExecutionByScriptId := func(scriptId string) []*executor.Execution {
		executions := make([]*executor.Execution, 0)
		for _, e := range service.executions.List() {
			if e.GetScriptInfo().ScriptID == scriptId {
				executions = append(executions, e)
			}
		}
		return executions
	}

	executions := findExecutionByScriptId(id)

	for _, e := range executions {
		if e.GetStatus() == "running" {
			return -1, ierrors.ScriptIsRunningError
		}
	}

	count, err := service.scriptRepo.Delete(id)

	if err != nil {
		return -1, ierrors.DeleteScriptDbError
	}

	return count, nil
}

func (service *ScriptService) MakeAndStartExecution(
	id string,
	command []string,
	environmentsId []string,
) (*executor.Execution, error) {
	script, err := service.scriptRepo.Get(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ierrors.ScriptNotFound
		}
		return nil, ierrors.GetScriptDbError
	}

	// 优先使用前端传递的覆盖参数，否则回退到脚本存储值。
	if command == nil {
		command = script.Command
	}
	if command == nil {
		command = make([]string, 0)
	}

	if environmentsId == nil {
		environmentsId = script.Environments
	}

	env, apiErr := service.resolveEnvironmentVars(environmentsId)
	if apiErr != nil {
		return nil, apiErr
	}

	execution := executor.NewExecution(
		script.ID,
		script.Name,
		command,
		script.WorkDir,
		env,
	)

	startErr := execution.Start(context.Background())

	if startErr != nil {
		return nil, ierrors.ExecuteScriptError
	}

	service.executions.Add(execution)

	return execution, nil
}

// GetExecution returns the tracked execution by id, or a 404 when unknown.
func (service *ScriptService) GetExecution(id string) (*executor.Execution, error) {
	execution, ok := service.executions.Get(id)
	if !ok {
		return nil, ierrors.ExecutionNotFound
	}
	return execution, nil
}

// ListExecutions returns an id/status snapshot of every tracked execution,
// ordered by start time. Records persist after the process exits until they
// are removed explicitly via KillExecution.
func (service *ScriptService) ListExecutions() (any, error) {
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

// KillExecution removes a tracked execution record by id. A still-running
// process is killed first so its handle is never lost mid-flight.
func (service *ScriptService) KillExecution(id string) (any, error) {
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

// resolveEnvironmentVars 展平所有的链式依赖，列表后续的 env 可以覆盖前面的同名变量。
// gen by ai
func (service *ScriptService) resolveEnvironmentVars(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return make([]string, 0), nil
	}

	values := make(map[string]string)
	paths := linkedhashset.New()

	for _, id := range ids {
		environment, err := service.environmentService.getEnvironment(id)

		if err != nil {
			return nil, err
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
