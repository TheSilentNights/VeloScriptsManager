package routers

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
)

type ScriptsRouter struct {
	scriptService      *services.ScriptService
	environmentService *services.EnvironmentService
	callerService      *services.CallerService
}

func NewScriptsRouter(scriptService *services.ScriptService, environmentService *services.EnvironmentService, callerService *services.CallerService) *ScriptsRouter {
	return &ScriptsRouter{
		scriptService:      scriptService,
		environmentService: environmentService,
		callerService:      callerService,
	}
}

func (router *ScriptsRouter) RegisterRoutes(engine *gin.RouterGroup) {
	scriptsGroup := engine.Group("/scripts")
	{
		scriptsGroup.GET("/", router.getStoredScripts)
		scriptsGroup.POST("/add", router.AddScript)
		scriptsGroup.PUT("/update", router.UpdateScript)
		scriptsGroup.DELETE("/delete", router.DeleteScript)
	}
}

func (router *ScriptsRouter) getStoredScripts(c *gin.Context) {
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

func (router *ScriptsRouter) AddScript(c *gin.Context) {
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

func (router *ScriptsRouter) DeleteScript(c *gin.Context) {
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

	count, apiErr := router.scriptService.DeleteScript(req.Id, services.GetExecutionProvider())
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

func (router *ScriptsRouter) UpdateScript(c *gin.Context) {
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
func (router *ScriptsRouter) ExecuteScript(c *gin.Context) {
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

	execution, apiErr := router.callerService.MakeAndStartExecution(
		req.Id,
		req.Command,
		req.EnvironmentsId,
		router.scriptService,
		router.environmentService,
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

func (scriptRouter *ScriptsRouter) GetScriptService() *services.ScriptService {
	return scriptRouter.scriptService
}
