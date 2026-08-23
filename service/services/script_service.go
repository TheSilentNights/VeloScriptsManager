package services

import (
	"github.com/gin-gonic/gin"
)

func (service *Service) ListScripts(c *gin.Context) {
	list, err := service.scriptRepo.List()

	if err != nil {
		c.Errors.JSON()
	}

	c.JSON(200, list)
}
