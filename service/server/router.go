package main

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/events"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/server/routers"
	"strings"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	scriptRouter      *routers.ScriptsRouter
	environmentRouter *routers.EnvironmentRouter
	serverController  *services.Server
}

func NewRouter(
	scriptService *services.ScriptService,
	environmentService *services.EnvironmentService,
	serverController *services.Server,
) *Router {
	return &Router{
		environmentRouter: routers.NewEnvironmentRouter(environmentService),
		scriptRouter:      routers.NewScriptsRouter(scriptService),
		serverController:  serverController,
	}
}

func (router *Router) RegisterRoutes(engine *gin.Engine) {

	engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return origin == "null" ||
				origin == "http://localhost:5173" ||
				origin == "http://127.0.0.1:5173" ||
				strings.HasPrefix(origin, "http://127.0.0.1:")
		},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/status", router.getStatus)

	api := engine.Group("/api/v1/")
	api.GET("/getExecutions", router.getExecutions)

	api.POST("/addScript", router.AddScript)
	api.POST("/updateScript", router.UpdateScript)
	api.POST("/deleteScript", router.DeleteScript)
	api.POST("/executeScript", router.ExecuteScript)

	api.POST("/deleteExecution", router.killExecution)
	api.POST("/stop", router.stopServer)

	api.POST("/registerFileChangeEvent", router.RegisterFileChangeEvent)

	api.GET("/getConfig", router.getConfig)
	api.POST("/updateConfig", router.updateConfig)
}

func (router *Router) getStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (router *Router) stopServer(c *gin.Context) {
	router.serverController.StopServer()
	c.JSON(200, gin.H{
		"message": "server is stopping",
	})
}

// getExecutions returns the id/status snapshot of all tracked executions.
func (router *Router) getExecutions(c *gin.Context) {
	result, apiErr := router.scriptService.ListExecutions()
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "list executions failed",
			"data":    apiErr.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    result,
	})

}

func (router *Router) killExecution(c *gin.Context) {
	req := &models.DeleteRequest{}

	if err := c.ShouldBind(req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	if len(req.Id) == 0 {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
		})
		return
	}

	execution, err := router.scriptRouter.GetScriptService().KillExecution(req.Id)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "kill execution failed",
			"data":    err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    execution,
	})

}

func (router *Router) RegisterFileChangeEvent(c *gin.Context) {
	req := &models.RegisterFileChangeEventRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
		})
		return
	}
	events.RegisterFileChangeEvent(req.Path, func() {
		router.scriptRouter.GetScriptService().MakeAndStartExecution(
			req.Id,
			req.Command,
			req.EnvironmentsId,
		)
	})
}

func (router *Router) getConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "success",
		"data":    configs.GetConfig(),
	})
}

func (router *Router) updateConfig(c *gin.Context) {
	req := &configs.Config{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	err := configs.SetConfig(*req)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "update config failed",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    req,
	})
}
