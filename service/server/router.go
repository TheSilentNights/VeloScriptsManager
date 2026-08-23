package main

import (
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
)

var shutdownChan = make(chan struct{})

type Router struct {
	service *services.Service
}

func NewRouter(service *services.Service) *Router {
	return &Router{service: service}
}

func (router *Router) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/status", router.getStatus)
	api := engine.Group("/api/v1/")
	api.GET("/getTaskLists", router.service.ListScripts)
	api.POST("/stop", router.stopServer)

}

func (router *Router) getStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (router *Router) stopServer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "server is stopping",
	})

	close(shutdownChan)
}

func (router *Router) getStoredScripts(c *gin.Context) {
	router.service.ListScripts(c)
}

func getRunningTaskList(c *gin.Context) {

}
