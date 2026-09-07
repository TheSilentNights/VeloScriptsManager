package services

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"

	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

var scriptService *ScriptService

type ScriptProvider interface {
	GetScript(id string) (*storage.Script, error)
}

type ScriptService struct {
	scriptRepo *storage.ScriptRepo
}

func NewScriptService(scriptRepo *storage.ScriptRepo) *ScriptService {
	scriptService = &ScriptService{
		scriptRepo: scriptRepo,
	}
	return scriptService
}

func GetScriptProvider() ScriptProvider {
	return scriptService
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

func (service *ScriptService) DeleteScript(id string, executionProvider ExecutionProvider) (int64, error) {

	findExecutionByScriptId := func(scriptId string) []*executor.Execution {
		executions := make([]*executor.Execution, 0)
		for _, e := range executionProvider.List() {
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

func (service *ScriptService) GetScript(id string) (*storage.Script, error) {
	return service.scriptRepo.Get(id)
}
