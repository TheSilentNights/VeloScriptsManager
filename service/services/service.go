package services

import (
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"sync"

	"github.com/gin-gonic/gin"
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

func (service *Service) StopServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "server is stopping",
	})

	service.shutdownOnce.Do(func() {
		close(service.GetShutdownSignalChan())
	})
}
