package services

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListEnvironments() ([]storage.Environment, error) {
	return service.environmentRepo.List()
}

func (service *Service) AddEnvironment(req *models.AddEnvironmentRequest) error {
	environment := storage.Environment{
		ID:   utils.GenerateEnvironmentId(),
		Name: req.Name,
		Type: req.Type,
		Path: req.Path,
	}

	if err := service.environmentRepo.Upsert(environment); err != nil {
		return errors.New(ierrors.DbError)
	}

	return nil
}

func (service *Service) DeleteEnvironment(id string) error {
	if err := service.environmentRepo.Delete(id); err != nil {
		return errors.New(ierrors.DbError)
	}
	return nil
}