package routers

import (
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
)

type ExecutionRouter struct {
	executionService *services.ExecutionService
	callerService    *services.CallerService
}

func NewExecutionRouter(executionService *services.ExecutionService, callerService *services.CallerService) *ExecutionRouter {
	return &ExecutionRouter{
		callerService:    callerService,
		executionService: executionService,
	}
}

func (router *ExecutionRouter) RegisterRoutes(engine *gin.RouterGroup) {
	api := engine.Group("/execution")
	api.GET("/", router.getExecutions)
	api.POST("/execute", router.startExecution)
	api.POST("/kill", router.killExecution)
}

func (router *ExecutionRouter) getExecutions(c *gin.Context) {
	result, apiErr := router.executionService.ListExecutions()
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

func (router *ExecutionRouter) startExecution(c *gin.Context) {
	req := &models.ExecuteScriptRequest{}

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

	_, err := services.GetCaller().MakeAndStartExecution(
		req.Id,
		req.Command,
		req.EnvironmentsId,
		services.GetScriptProvider(),
		services.GetEnvironmentProvider(),
	)

	if err != nil {
		c.JSON(500, gin.H{
			"message": "start execution failed",
			"data":    err.Error(),
		})
		return
	}

}
func (router *ExecutionRouter) killExecution(c *gin.Context) {
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

	execution, err := router.executionService.KillExecution(req.Id)
	if err != nil {
		switch err {
		case ierrors.ExecutionNotRunningError:
			c.JSON(400, gin.H{
				"message": "execution not running",
			})
			return
		default:
			c.JSON(500, gin.H{
				"message": "kill execution failed",
				"data":    err.Error(),
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    execution,
	})

}
