package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListEnvironments() ([]storage.Environment, *models.ApiError) {
	list, err := service.environmentRepo.List()
	if err != nil {
		return nil, ierrors.DbError
	}
	return list, nil
}

func (service *Service) AddEnvironment(req *models.AddEnvironmentRequest) *models.ApiError {
	environment := storage.Environment{
		ID:   utils.GenerateEnvironmentId(),
		Name: req.Name,
		Type: req.Type,
		Path: req.Path,
	}

	if err := service.environmentRepo.Upsert(environment); err != nil {
		return ierrors.DbError
	}

	return nil
}

func (service *Service) DeleteEnvironment(id string) *models.ApiError {
	if err := service.environmentRepo.Delete(id); err != nil {
		return ierrors.DbError
	}
	return nil
}