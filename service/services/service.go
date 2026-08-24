package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"
)

type Service struct {
	scriptRepo      *storage.ScriptRepo
	environmentRepo *storage.EnvironmentRepo
	executions      *executionManager

	shutdownSignalChan chan struct{}
	shutdownOnce       sync.Once
}

func NewService(
	scriptRepo *storage.ScriptRepo,
	environmentRepo *storage.EnvironmentRepo,
) *Service {
	return &Service{
		scriptRepo:         scriptRepo,
		environmentRepo:    environmentRepo,
		executions:         newExecutionManager(),
		shutdownSignalChan: make(chan struct{}),
		shutdownOnce:       sync.Once{},
	}
}

func (service *Service) GetShutdownSignalChan() chan struct{} {
	return service.shutdownSignalChan
}

func (service *Service) StopServer() {
	service.shutdownOnce.Do(func() {
		close(service.GetShutdownSignalChan())
	})
}
