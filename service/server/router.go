package main

import (
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
)

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
	api.POST("/addScript", router.AddScript)
	api.POST("/stop", router.stopServer)

}

func (router *Router) getStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (router *Router) stopServer(c *gin.Context) {
	router.service.StopServer(c)
}

func (router *Router) getStoredScripts(c *gin.Context) {
	router.service.ListScripts(c)
}

func (router *Router) AddScript(c *gin.Context) {
	req := &models.AddScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "arguments are not valid",
		})
		return
	}

	err := router.service.AddScript(req)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
}
