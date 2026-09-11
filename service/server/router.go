package main

import (
	"github/TheSilentNights/VeloScriptsManager/service/server/routers"
	"strings"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	scriptRouter      *routers.ScriptsRouter
	environmentRouter *routers.EnvironmentRouter
	executionRouter   *routers.ExecutionRouter
	eventRouter       *routers.EventRouter
	serverController  *services.Server
}

func NewRouter(
	scriptService *services.ScriptService,
	environmentService *services.EnvironmentService,
	serverController *services.Server,
	executionService *services.ExecutionService,
	callerService *services.CallerService,
) *Router {
	return &Router{
		environmentRouter: routers.NewEnvironmentRouter(environmentService),
		scriptRouter:      routers.NewScriptsRouter(scriptService, environmentService, callerService),
		executionRouter:   routers.NewExecutionRouter(executionService, callerService),
		eventRouter:       routers.NewEventRouter(callerService),
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
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/status", router.getStatus)

	api := engine.Group("/api/v1/")

	router.scriptRouter.RegisterRoutes(api)
	router.environmentRouter.RegisterRoutes(api)
	router.executionRouter.RegisterRoutes(api)
	router.eventRouter.RegisterRoutes(api)

	api.POST("/stop", router.stopServer)
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

	err := configs.SetConfig(req)
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
