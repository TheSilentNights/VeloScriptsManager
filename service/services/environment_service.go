package services

import (
	"errors"

	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

func (service *Service) ListEnvironments() (*models.Result, *models.ApiError) {
	list, err := service.environmentRepo.List()
	if err != nil {
		return nil, models.NewApiError(500, "list environments fail", err.Error())
	}
	return models.NewResult(list), nil
}

func (service *Service) AddEnvironment(req *models.AddEnvironmentRequest) (*models.Result, *models.ApiError) {
	environment := storage.Environment{
		ID:       utils.GenerateEnvironmentId(),
		Name:     req.Name,
		Type:     req.Type,
		Path:     req.Path,
		Env:      req.Env,
		Children: req.Children,
	}

	if err := service.environmentRepo.Upsert(environment); err != nil {
		return nil, models.NewApiError(500, "add environment fail", err.Error())
	}

	return models.NewResultWithMessage("environment added", nil), nil
}

func (service *Service) DeleteEnvironment(id string) (*models.Result, *models.ApiError) {
	if err := service.environmentRepo.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "environment not found", id)
		}
		return nil, models.NewApiError(500, "delete environment fail", err.Error())
	}
	return models.NewResultWithMessage("environment deleted", nil), nil
}
