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
	api.GET("/getTaskLists", router.getStoredScripts)
	api.POST("/addScript", router.AddScript)
	api.POST("/deleteScript", router.DeleteScript)
	api.GET("/getEnvironments", router.getStoredEnvironments)
	api.POST("/addEnvironment", router.AddEnvironment)
	api.POST("/deleteEnvironment", router.DeleteEnvironment)
	api.POST("/stop", router.stopServer)

}

func (router *Router) getStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (router *Router) stopServer(c *gin.Context) {
	router.service.StopServer()
	c.JSON(200, gin.H{
		"message": "server is stopping",
	})
}

func (router *Router) getStoredScripts(c *gin.Context) {
	list, err := router.service.ListScripts()
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, list)
}

func (router *Router) DeleteScript(c *gin.Context) {
	req := &models.DeleteScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "arguments are not valid",
		})
		return
	}

	if req.ID == "" {
		c.JSON(400, gin.H{
			"error": "id is required",
		})
		return
	}

	err := router.service.DeleteScript(req.ID)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "script deleted",
	})
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	list, err := router.service.ListEnvironments()
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, list)
}

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "arguments are not valid",
		})
		return
	}

	err := router.service.AddEnvironment(req)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "environment added",
	})
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
	req := &models.DeleteEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "arguments are not valid",
		})
		return
	}

	if req.ID == "" {
		c.JSON(400, gin.H{
			"error": "id is required",
		})
		return
	}

	err := router.service.DeleteEnvironment(req.ID)

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "environment deleted",
	})
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
