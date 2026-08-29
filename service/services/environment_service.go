package services

import (
	"errors"

	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
)

type EnvironmentService struct {
	environmentRepo *storage.EnvironmentRepo
}

func NewEnvironmentService(environmentRepo *storage.EnvironmentRepo) *EnvironmentService {
	return &EnvironmentService{environmentRepo: environmentRepo}
}

func (service *EnvironmentService) ListEnvironments() (*models.Result, *models.ApiError) {
	list, err := service.environmentRepo.List()
	if err != nil {
		return nil, models.NewApiError(500, "list environments fail", err.Error())
	}
	return models.NewResult(list), nil
}

func (service *EnvironmentService) AddEnvironment(req *models.AddEnvironmentRequest) (*models.Result, *models.ApiError) {
	environment := storage.Environment{
		ID:    utils.GenerateEnvironmentId(),
		Name:  req.Name,
		Paths: req.Paths,
		Env:   req.Env,
	}

	if err := service.environmentRepo.Upsert(environment); err != nil {
		return nil, models.NewApiError(500, "add environment fail", err.Error())
	}

	return models.NewResultWithMessage("environment added", nil), nil
}

func (service *EnvironmentService) UpdateEnvironment(req *models.UpdateEnvironmentRequest) (*models.Result, *models.ApiError) {
	if req.Id == "" {
		return nil, models.NewApiError(400, "invalid arguments", "id cannot be empty")
	}

	if _, err := service.environmentRepo.Get(req.Id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "environment not found", req.Id)
		}
		return nil, models.NewApiError(500, "get environment fail", err.Error())
	}

	environment := storage.Environment{
		ID:    req.Id,
		Name:  req.Name,
		Paths: req.Paths,
		Env:   req.Env,
	}

	if err := service.environmentRepo.Upsert(environment); err != nil {
		return nil, models.NewApiError(500, "update environment fail", err.Error())
	}

	return models.NewResultWithMessage("environment updated", nil), nil
}

func (service *EnvironmentService) DeleteEnvironment(id string) (*models.Result, *models.ApiError) {
	if err := service.environmentRepo.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, models.NewApiError(404, "environment not found", id)
		}
		return nil, models.NewApiError(500, "delete environment fail", err.Error())
	}
	return models.NewResultWithMessage("environment deleted", nil), nil
}

func (service *EnvironmentService) getEnvironment(id string) (*storage.Environment, error) {
	result, err := service.environmentRepo.Get(id)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
	}

	return result, nil
}
