package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListScripts() ([]storage.Script, *models.ApiError) {
	list, err := service.scriptRepo.List()
	if err != nil {
		return nil, ierrors.DbError
	}
	return list, nil
}

func (service *Service) AddScript(req *models.AddScriptRequest) *models.ApiError {

	script := storage.Script{
		ID:      utils.GenerateScriptId(),
		Name:    req.Name,
		Command: req.Command,
		WorkDir: req.WorkDir,
		Runner:  req.Runner,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return ierrors.DbError
	}

	return nil
}

func (service *Service) DeleteScript(id string) *models.ApiError {
	if err := service.scriptRepo.Delete(id); err != nil {
		return ierrors.DbError
	}
	return nil
}