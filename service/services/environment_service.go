package services

import (
	"database/sql"
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"

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

func (service *EnvironmentService) ListEnvironments() (any, error) {
	list, err := service.environmentRepo.List()
	if err != nil {
		return nil, ierrors.ListEnvironmentsError
	}
	return list, nil
}

func (service *EnvironmentService) AddEnvironment(req *models.AddEnvironmentRequest) (any, error) {
	environment := storage.Environment{
		ID:    utils.GenerateEnvironmentId(),
		Name:  req.Name,
		Paths: req.Paths,
		Env:   req.Env,
	}

	count, err := service.environmentRepo.Insert(environment)
	if err != nil {
		return nil, ierrors.AddEnvironmentDbError
	}

	return count, nil
}

func (service *EnvironmentService) UpdateEnvironment(req *models.UpdateEnvironmentRequest) (any, error) {

	environment := storage.Environment{
		ID:    req.Id,
		Name:  req.Name,
		Paths: req.Paths,
		Env:   req.Env,
	}

	count, err := service.environmentRepo.Update(environment)
	if err != nil {
		return nil, ierrors.UpdateEnvironmentDbError
	}

	return count, nil
}

func (service *EnvironmentService) DeleteEnvironment(id string) (any, error) {
	count, err := service.environmentRepo.Delete(id)
	if err != nil {
		return nil, ierrors.DeleteEnvironmentDbError
	}
	return count, nil
}

func (service *EnvironmentService) getEnvironment(id string) (*storage.Environment, error) {
	result, err := service.environmentRepo.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ierrors.EnvironmentNotFound
		default:
			return nil, ierrors.GetEnvironmentDbError
		}
	}

	return result, nil
}
