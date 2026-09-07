package services

import (
	"database/sql"
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"sort"
	"strings"

	"github.com/emirpasic/gods/sets/linkedhashset"
)

type Caller interface {
	MakeAndStartExecution(
		scriptId string,
		command []string,
		environmentsId []string,
		scriptProvider ScriptProvider,
		environmentProvider EnvironmentProvider,
	) (*executor.Execution, error)
}

type CallerService struct {
	executionService *ExecutionService
}

var callerService *CallerService

func NewCallerService(executionService *ExecutionService) *CallerService {
	callerService = &CallerService{
		executionService: executionService,
	}
	return callerService
}

func GetCaller() Caller {
	return callerService
}

func (service *CallerService) MakeAndStartExecution(
	scriptId string,
	command []string,
	environmentsId []string,
	scriptProvider ScriptProvider,
	environmentProvider EnvironmentProvider,
) (*executor.Execution, error) {
	script, err := scriptProvider.GetScript(scriptId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ierrors.ScriptNotFound
		}
		return nil, ierrors.GetScriptDbError
	}

	if len(command) == 0 {
		command = script.Command
	}

	if len(environmentsId) == 0 {
		environmentsId = script.Environments
	}

	env, apiErr := resolveEnvironmentVars(environmentsId, environmentProvider)
	if apiErr != nil {
		return nil, apiErr
	}

	return service.executionService.MakeAndStartExecution(
		scriptId,
		script.Name,
		command,
		script.WorkDir,
		env,
	)
}

func resolveEnvironmentVars(ids []string, environmentProvider EnvironmentProvider) ([]string, error) {
	if len(ids) == 0 {
		return make([]string, 0), nil
	}

	values := make(map[string]string)
	paths := linkedhashset.New()

	for _, id := range ids {
		environment, err := environmentProvider.GetEnvironment(id)

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
