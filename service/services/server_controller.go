package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"
)

type Server struct {
	environmentRepo    *storage.EnvironmentRepo
	executionManager   *executor.ExecutionManager
	shutdownSignalChan chan struct{}
	shutdownOnce       sync.Once
}

func NewServerController(
	manager *executor.ExecutionManager,
) *Server {
	return &Server{
		executionManager:   manager,
		shutdownSignalChan: make(chan struct{}),
		shutdownOnce:       sync.Once{},
	}
}

func (service *Server) GetShutdownSignalChan() chan struct{} {
	return service.shutdownSignalChan
}

// StopServer terminates all running script processes and then signals the
// main goroutine to begin the graceful HTTP shutdown.
func (service *Server) StopServer() {
	service.shutdownOnce.Do(func() {
		service.executionManager.KillRunning()
		close(service.GetShutdownSignalChan())
	})
}
