package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"
)

type Service struct {
	scriptRepo      *storage.ScriptRepo
	environmentRepo *storage.EnvironmentRepo

	shutdownSignalChan chan struct{}
	shutdownOnce       sync.Once
}

func NewService(
	scriptRepo *storage.ScriptRepo,
	environmentRepo *storage.EnvironmentRepo,
) *Service {
	return &Service{scriptRepo, environmentRepo, make(chan struct{}), sync.Once{}}
}

func (service *Service) GetShutdownSignalChan() chan struct{} {
	return service.shutdownSignalChan
}

func (service *Service) StopServer() {
	service.shutdownOnce.Do(func() {
		close(service.GetShutdownSignalChan())
	})
}
