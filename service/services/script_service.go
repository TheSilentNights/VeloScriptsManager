package services

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListScripts() ([]storage.Script, error) {
	return service.scriptRepo.List()
}

func (service *Service) AddScript(req *models.AddScriptRequest) error {

	script := storage.Script{
		ID:      utils.GenerateScriptId(),
		Name:    req.Name,
		Command: req.Command,
		WorkDir: req.WorkDir,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return errors.New(ierrors.DbError)
	}

	return nil
}

func (service *Service) DeleteScript(id string) error {
	if err := service.scriptRepo.Delete(id); err != nil {
		return errors.New(ierrors.DbError)
	}
	return nil
}
