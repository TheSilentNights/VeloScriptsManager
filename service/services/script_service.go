package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListScripts() (*models.Result, *models.ApiError) {
	list, err := service.scriptRepo.List()
	if err != nil {
		return nil, ierrors.DbError
	}
	return models.NewResult(list), nil
}

func (service *Service) AddScript(req *models.AddScriptRequest) (*models.Result, *models.ApiError) {

	script := storage.Script{
		ID:      utils.GenerateScriptId(),
		Name:    req.Name,
		Command: req.Command,
		WorkDir: req.WorkDir,
		Runner:  req.Runner,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return nil, ierrors.DbError
	}

	return models.NewResultWithMessage("script added", nil), nil
}

func (service *Service) DeleteScript(id string) (*models.Result, *models.ApiError) {
	if err := service.scriptRepo.Delete(id); err != nil {
		return nil, ierrors.DbError
	}
	return models.NewResultWithMessage("script deleted", nil), nil
}