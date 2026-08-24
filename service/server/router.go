package main

import (
	"fmt"
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

func writeResult(c *gin.Context, result *models.Result) {
	c.JSON(result.Code, result)
}

func writeError(c *gin.Context, apiErr *models.ApiError) {
	c.JSON(apiErr.Code, apiErr)
}

func (router *Router) getStoredScripts(c *gin.Context) {
	result, apiErr := router.service.ListScripts()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) AddScript(c *gin.Context) {
	req := &models.AddScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	result, apiErr := router.service.AddScript(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteScript(c *gin.Context) {
	router.handleDelete(c, router.service.DeleteScript)
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	result, apiErr := router.service.ListEnvironments()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	result, apiErr := router.service.AddEnvironment(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
	router.handleDelete(c, router.service.DeleteEnvironment)
}

func (router *Router) handleDelete(c *gin.Context, deleteFn func(id string) (*models.Result, *models.ApiError)) {
	var req models.DeleteRequest

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	if req.Id == "" {
		writeError(c, models.NewApiError(400, "invalid arguments", fmt.Sprintf("id cannot be empty")))
		return
	}

	result, apiErr := deleteFn(req.Id)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}
