package main

import (
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
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

// writeError 将服务层返回的业务错误（与 gin 解耦）写入响应
func writeError(c *gin.Context, apiErr *models.ApiError) {
	c.JSON(apiErr.Status, apiErr)
}

func (router *Router) getStoredScripts(c *gin.Context) {
	list, apiErr := router.service.ListScripts()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	c.JSON(200, list)
}

func (router *Router) AddScript(c *gin.Context) {
	req := &models.AddScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, ierrors.InvalidArgument)
		return
	}

	if apiErr := router.service.AddScript(req); apiErr != nil {
		writeError(c, apiErr)
		return
	}
}

func (router *Router) DeleteScript(c *gin.Context) {
	req := &models.DeleteScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, ierrors.InvalidArgument)
		return
	}

	if req.ID == "" {
		writeError(c, ierrors.IdRequired)
		return
	}

	if apiErr := router.service.DeleteScript(req.ID); apiErr != nil {
		writeError(c, apiErr)
		return
	}

	c.JSON(200, gin.H{
		"message": "script deleted",
	})
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	list, apiErr := router.service.ListEnvironments()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	c.JSON(200, list)
}

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, ierrors.InvalidArgument)
		return
	}

	if apiErr := router.service.AddEnvironment(req); apiErr != nil {
		writeError(c, apiErr)
		return
	}

	c.JSON(200, gin.H{
		"message": "environment added",
	})
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
	req := &models.DeleteEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, ierrors.InvalidArgument)
		return
	}

	if req.ID == "" {
		writeError(c, ierrors.IdRequired)
		return
	}

	if apiErr := router.service.DeleteEnvironment(req.ID); apiErr != nil {
		writeError(c, apiErr)
		return
	}

	c.JSON(200, gin.H{
		"message": "environment deleted",
	})
}