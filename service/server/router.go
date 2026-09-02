package main

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"strings"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	scriptService      *services.ScriptService
	environmentService *services.EnvironmentService
	serverController   *services.Server
}

func NewRouter(
	scriptService *services.ScriptService,
	environmentService *services.EnvironmentService,
	serverController *services.Server,
) *Router {
	return &Router{
		scriptService:      scriptService,
		environmentService: environmentService,
		serverController:   serverController,
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
	api.GET("/getStoredScripts", router.getStoredScripts)
	api.GET("/getEnvironments", router.getStoredEnvironments)
	api.GET("/getExecutions", router.getExecutions)

	api.POST("/addScript", router.AddScript)
	api.POST("/updateScript", router.UpdateScript)
	api.POST("/deleteScript", router.DeleteScript)
	api.POST("/executeScript", router.ExecuteScript)
	api.POST("/addEnvironment", router.AddEnvironment)
	api.POST("/updateEnvironment", router.UpdateEnvironment)
	api.POST("/deleteEnvironment", router.DeleteEnvironment)
	api.POST("/deleteExecution", router.killExecution)
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

func (router *Router) getStoredScripts(c *gin.Context) {
	result, apiErr := router.scriptService.ListScripts()

	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "list scripts failed",
			"data":    apiErr.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    result,
	})
}

func (router *Router) AddScript(c *gin.Context) {
	req := &models.AddScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	if len(req.Name) == 0 {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    req,
		})
		return
	}

	result, apiErr := router.scriptService.AddScript(req)
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "add script failed",
			"data":    apiErr.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    result,
	})
}

func (router *Router) DeleteScript(c *gin.Context) {
	var req models.DeleteRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	if req.Id == "" {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
		})
		return
	}

	count, apiErr := router.scriptService.DeleteScript(req.Id)
	if apiErr != nil {
		switch {
		case errors.Is(apiErr, ierrors.ScriptIsRunningError):
			c.JSON(400, gin.H{
				"message": "script is running",
				"data":    apiErr.Error(),
			})
		case errors.Is(apiErr, ierrors.DeleteScriptDbError):
			c.JSON(500, gin.H{
				"message": "delete script failed",
				"data":    apiErr.Error(),
			})
		default:
			c.JSON(500, gin.H{
				"message": "delete script failed",
				"data":    apiErr.Error(),
			})
		}
		return
	}

	if count == 0 {
		c.JSON(404, gin.H{
			"message": "script not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    req,
	})
}

func (router *Router) UpdateScript(c *gin.Context) {
	req := &models.UpdateScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	count, apiErr := router.scriptService.UpdateScript(req)
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "update script failed",
			"data":    apiErr.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(404, gin.H{
			"message": "script not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    req,
	})
}

// ExecuteScript starts the script identified by the request id asynchronously
// and returns the execution id immediately.
func (router *Router) ExecuteScript(c *gin.Context) {
	var req models.ExecuteScriptRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	if req.Id == "" {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    req,
		})
		return
	}

	execution, apiErr := router.scriptService.MakeAndStartExecution(
		req.Id,
		req.Command,
		req.EnvironmentsId,
	)

	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "execute script failed",
			"data":    apiErr.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"executionId": execution.GetExecutionId(),
			"scriptId":    execution.GetScriptInfo().ScriptID,
			"name":        execution.GetScriptInfo().Name,
		},
	})
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	result, apiErr := router.environmentService.ListEnvironments()
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "list environments failed",
			"data":    apiErr.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    result,
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

	execution, err := router.scriptService.KillExecution(req.Id)
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

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	if len(req.Name) == 0 {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
		})
		return
	}

	count, apiErr := router.environmentService.AddEnvironment(req)
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "add environment failed",
			"data":    apiErr.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    count,
	})
}

func (router *Router) UpdateEnvironment(c *gin.Context) {
	req := &models.UpdateEnvironmentRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}

	count, apiErr := router.environmentService.UpdateEnvironment(req)
	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "update environment failed",
			"data":    apiErr.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(404, gin.H{
			"message": "environment not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    count,
	})
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
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
			"data":    req,
		})
		return
	}

	execution, apiErr := router.environmentService.DeleteEnvironment(req.Id)

	if apiErr != nil {
		c.JSON(500, gin.H{
			"message": "delete environment failed",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    execution,
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
