package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"
)

type Service struct {
	scriptRepo         *storage.ScriptRepo
	shutdownSignalChan chan struct{}
	shutdownOnce       sync.Once
}

func NewService(repo *storage.ScriptRepo) *Service {
	return &Service{repo, make(chan struct{}), sync.Once{}}
}

func (service *Service) GetShutdownSignalChan() chan struct{} {
	return service.shutdownSignalChan
}

func (service *Service) StopServer() {
	service.shutdownOnce.Do(func() {
		close(service.GetShutdownSignalChan())
	})
}
