package main

import (
	"fmt"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var trustedHost = []string{
	"localhost",
	"127.0.0.1",
}
var wsUpgrader = websocket.Upgrader{
	// The management UI is served from a different origin; accept all.
	CheckOrigin: func(r *http.Request) bool {
		if slices.Contains(trustedHost, r.Host) {
			return true
		}

		return false
	},
}

type Router struct {
	scriptService      *services.ScriptService
	environmentService *services.EnvironmentService
	wsService          *services.WsService
	serverController   *services.Server
}

func NewRouter(
	scriptService *services.ScriptService,
	environmentService *services.EnvironmentService,
	wsService *services.WsService,
	serverController *services.Server,
) *Router {
	return &Router{
		scriptService:      scriptService,
		environmentService: environmentService,
		wsService:          wsService,
		serverController:   serverController,
	}
}

func (router *Router) RegisterRoutes(engine *gin.Engine) {

	engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:5173" ||
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
	api.GET("/execute/attach", router.attachExecution)
	api.GET("/getEnvironments", router.getStoredEnvironments)
	api.GET("/getExecutions", router.getExecutions)

	api.POST("/addScript", router.AddScript)
	api.POST("/updateScript", router.UpdateScript)
	api.POST("/deleteScript", router.DeleteScript)
	api.POST("/executeScript", router.ExecuteScript)
	api.POST("/addEnvironment", router.AddEnvironment)
	api.POST("/deleteEnvironment", router.DeleteEnvironment)
	api.POST("/deleteExecution", router.DeleteExecution)
	api.POST("/stop", router.stopServer)
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

func writeResult(c *gin.Context, result *models.Result) {
	c.JSON(result.Code, result)
}

func writeError(c *gin.Context, apiErr *models.ApiError) {
	c.JSON(apiErr.Code, apiErr)
}

func (router *Router) getStoredScripts(c *gin.Context) {
	result, apiErr := router.scriptService.ListScripts()
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

	result, apiErr := router.scriptService.AddScript(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteScript(c *gin.Context) {
	router.handleDelete(c, router.scriptService.DeleteScript)
}

func (router *Router) UpdateScript(c *gin.Context) {
	req := &models.UpdateScriptRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	result, apiErr := router.scriptService.UpdateScript(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

// ExecuteScript starts the script identified by the request id asynchronously
// and returns the execution id immediately. The caller can then attach to the
// running process stdio via GET /api/v1/execute/attach?executionId=...
func (router *Router) ExecuteScript(c *gin.Context) {
	var req models.ExecuteScriptRequest

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	if req.Id == "" {
		writeError(c, models.NewApiError(400, "invalid arguments", "id cannot be empty"))
		return
	}

	execution, apiErr := router.scriptService.StartExecution(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}

	writeResult(c, models.NewResultWithMessage("script started", gin.H{
		"executionId": execution.ID,
		"scriptId":    execution.ScriptID,
		"name":        execution.Name,
	}))
}

// attachExecution upgrades the connection to a WebSocket and bridges it to the
// stdio of a running execution.
//
// Server -> client frames (models.WsServerFrame):
//
//	{"type":"output","data":"<base64 chunk>","dropped":0}
//	{"type":"exit","code":0,"error":""}
//
// The server sends ping control frames every wsPingPeriod; clients must reply
// with pong (gorilla does this automatically). The server closes the socket
// right after the exit frame.
//
// Client -> server frames (models.WsClientFrame):
//
//	{"type":"stdin","data":"<base64 bytes>"}
//	{"type":"close_stdin"}
//	{"type":"kill"}
func (router *Router) attachExecution(c *gin.Context) {
	executionId := c.Query("executionId")
	if executionId == "" {
		writeError(c, models.NewApiError(400, "invalid arguments", "executionId cannot be empty"))
		return
	}

	execution, apiErr := router.scriptService.GetExecution(executionId)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	router.wsService.StreamExecution(conn, execution)
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	result, apiErr := router.environmentService.ListEnvironments()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

// getExecutions returns the id/status snapshot of all tracked executions.
func (router *Router) getExecutions(c *gin.Context) {
	result, apiErr := router.scriptService.ListExecutions()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteExecution(c *gin.Context) {
	router.handleDelete(c, router.scriptService.DeleteExecution)
}

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	result, apiErr := router.environmentService.AddEnvironment(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
	router.handleDelete(c, router.environmentService.DeleteEnvironment)
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
