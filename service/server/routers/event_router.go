package routers

import (
	"github/TheSilentNights/VeloScriptsManager/service/events"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
)

type EventRouter struct {
	callerService *services.CallerService
}

func NewEventRouter(callerService *services.CallerService) *EventRouter {
	return &EventRouter{
		callerService: callerService,
	}
}

func (router *EventRouter) RegisterRoutes(engine *gin.RouterGroup) {
	api := engine.Group("/event")
	api.POST("/registerFileChangeEvent", router.RegisterFileChangeEvent)
}

func (router *EventRouter) RegisterFileChangeEvent(c *gin.Context) {
	req := &models.RegisterFileChangeEventRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}
	events.RegisterFileChangeEvent(req.Path, func() {
		router.callerService.MakeAndStartExecution(
			req.Id,
			req.Command,
			req.EnvironmentsId,
			services.GetScriptProvider(),
			services.GetEnvironmentProvider(),
		)
	})
}
