package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"
)

type Server struct {
	environmentRepo    *storage.EnvironmentRepo
	executionManager   *ExecutionManager
	shutdownSignalChan chan struct{}
	shutdownOnce       sync.Once
}

func NewServerController(
	manager *ExecutionManager,
) *Server {
	return &Server{
		ExecutionManager:   manager,
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
		service.executionManager.killRunning()
		close(service.GetShutdownSignalChan())
	})
}
