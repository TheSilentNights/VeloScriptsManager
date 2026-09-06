package routers

import (
	"github.com/gin-gonic/gin"

	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"
)

type EnvironmentRouter struct {
	environmentService *services.EnvironmentService
}

func NewEnvironmentRouter(environmentService *services.EnvironmentService) *EnvironmentRouter {
	return &EnvironmentRouter{
		environmentService: environmentService,
	}
}

func (router *EnvironmentRouter) RegisterRoutes(engine *gin.Engine) {
	environmentGroup := engine.Group("/environments")
	{
		environmentGroup.GET("/", router.getStoredEnvironments)
		environmentGroup.POST("/add", router.AddEnvironment)
		environmentGroup.PUT("/update", router.UpdateEnvironment)
		environmentGroup.DELETE("/delete", router.DeleteEnvironment)
	}

}

func (router *EnvironmentRouter) getStoredEnvironments(c *gin.Context) {
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

func (router *EnvironmentRouter) AddEnvironment(c *gin.Context) {
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

func (router *EnvironmentRouter) UpdateEnvironment(c *gin.Context) {
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

func (router *EnvironmentRouter) DeleteEnvironment(c *gin.Context) {
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
