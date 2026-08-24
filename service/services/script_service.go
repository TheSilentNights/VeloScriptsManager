package services

import (
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
