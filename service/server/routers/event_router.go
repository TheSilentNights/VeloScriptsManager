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
	api.GET("/", router.getEvents)
	api.POST("/registerFileChangeEvent", router.registerFileChangeEvent)
	api.POST("/registerTimeEvent", router.registerTimeEvent)
}

func (router *EventRouter) getEvents(c *gin.Context) {
	fileChangeEventRegistry := events.GetFileChangeEventRegistry()
	timeEventRegistry := events.GetTimeEventRegistry()
	eventInfo := make([]models.EventInfoResponse, 0)
	for k, v := range fileChangeEventRegistry {
		eventInfo = append(eventInfo, models.EventInfoResponse{
			EventId:   k,
			Type:      "fileChangeEvent",
			EventData: v,
		})
	}
	for k, v := range timeEventRegistry {
		eventInfo = append(eventInfo, models.EventInfoResponse{
			EventId:   k,
			Type:      "timeEvent",
			EventData: v,
		})
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    eventInfo,
	})
}

func (router *EventRouter) registerFileChangeEvent(c *gin.Context) {
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

func (router *EventRouter) registerTimeEvent(c *gin.Context) {
	req := &models.RegisterTimeEventRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "invalid arguments",
			"data":    err.Error(),
		})
		return
	}
	events.RegisterTimeEvent(req.Interval, req.Repeat, func() {
		router.callerService.MakeAndStartExecution(
			req.Id,
			req.Command,
			req.EnvironmentsId,
			services.GetScriptProvider(),
			services.GetEnvironmentProvider(),
		)
	})
}
