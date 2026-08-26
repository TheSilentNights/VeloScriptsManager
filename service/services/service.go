package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"

	"github.com/gorilla/websocket"
)

type Service struct {
	scriptRepo      *storage.ScriptRepo
	environmentRepo *storage.EnvironmentRepo
	executions      *executionManager
	wsService       *WsStore

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
		wsService:          &WsStore{wsConns: make(map[*websocket.Conn]struct{})},
	}
}

func (service *Service) GetShutdownSignalChan() chan struct{} {
	return service.shutdownSignalChan
}

// StopServer terminates all running script processes and then signals the
// main goroutine to begin the graceful HTTP shutdown.
func (service *Service) StopServer() {
	service.shutdownOnce.Do(func() {
		service.executions.killRunning()
		close(service.GetShutdownSignalChan())
	})
}
