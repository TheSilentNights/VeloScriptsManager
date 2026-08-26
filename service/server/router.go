package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

var wsUpgrader = websocket.Upgrader{
	// The management UI is served from a different origin; accept all.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Router struct {
	service *services.Service

	wsMu    sync.Mutex
	wsConns map[*websocket.Conn]struct{}
}

func NewRouter(service *services.Service) *Router {
	return &Router{
		service: service,
		wsConns: make(map[*websocket.Conn]struct{}),
	}
}

// CloseWebSockets force-closes every attached execution WebSocket so a graceful
// shutdown is not blocked by long-lived connections.
func (router *Router) CloseWebSockets() {
	router.wsMu.Lock()
	defer router.wsMu.Unlock()
	for conn := range router.wsConns {
		_ = conn.Close()
	}
}

func (router *Router) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/status", router.getStatus)

	api := engine.Group("/api/v1/")
	api.GET("/getStoredScripts", router.getStoredScripts)
	api.GET("/execute/attach", router.attachExecution)
	api.GET("/getEnvironments", router.getStoredEnvironments)
	api.GET("/getExecutions", router.getExecutions)

	api.POST("/addScript", router.AddScript)
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
	router.service.StopServer()
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
	result, apiErr := router.service.ListScripts()
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

	result, apiErr := router.service.AddScript(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteScript(c *gin.Context) {
	router.handleDelete(c, router.service.DeleteScript)
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

	execution, apiErr := router.service.StartExecution(req.Id)
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

	execution, apiErr := router.service.GetExecution(executionId)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	router.streamExecution(conn, execution)
}

func (router *Router) streamExecution(conn *websocket.Conn, execution *models.Execution) {
	router.wsMu.Lock()
	router.wsConns[conn] = struct{}{}
	router.wsMu.Unlock()

	defer func() {
		router.wsMu.Lock()
		delete(router.wsConns, conn)
		router.wsMu.Unlock()
		_ = conn.Close()
	}()

	process := execution.Process()
	if process == nil {
		_ = conn.WriteJSON(models.WsServerFrame{
			Type:  "exit",
			Code:  execution.ExitCode(),
			Error: execution.Error(),
		})
		return
	}

	stop := make(chan struct{})
	writerDone := make(chan struct{})

	// Writer goroutine: forward output chunks and the final exit frame.
	go func() {
		defer close(writerDone)
		ch := process.Subscribe()
		pingTicker := time.NewTicker(wsPingPeriod)
		defer pingTicker.Stop()
		for {
			select {
			case chunk, ok := <-ch:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if !ok {
					// Process finished: report the exit.
					_ = conn.WriteJSON(models.WsServerFrame{
						Type:  "exit",
						Code:  process.ExitCode(),
						Error: errorString(process.Err()),
					})
					return
				}
				if err := conn.WriteJSON(models.WsServerFrame{
					Type:    "output",
					Data:    base64.StdEncoding.EncodeToString(chunk.Data),
					Dropped: chunk.Dropped,
				}); err != nil {
					return
				}
			case <-pingTicker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-stop:
				return
			}
		}
	}()

	// Reader loop: forward client frames to the process stdio.
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	for {
		var frame models.WsClientFrame
		if err := conn.ReadJSON(&frame); err != nil {
			break
		}

		switch frame.Type {
		case "stdin":
			data, err := base64.StdEncoding.DecodeString(frame.Data)
			if err != nil {
				continue
			}
			_, _ = process.WriteStdin(data)
		case "close_stdin":
			_ = process.CloseStdin()
		case "kill":
			_ = process.Kill()
		}
	}

	close(stop)
	<-writerDone
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (router *Router) getStoredEnvironments(c *gin.Context) {
	result, apiErr := router.service.ListEnvironments()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

// getExecutions returns the id/status snapshot of all tracked executions.
func (router *Router) getExecutions(c *gin.Context) {
	result, apiErr := router.service.ListExecutions()
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteExecution(c *gin.Context) {
	router.handleDelete(c, router.service.DeleteExecution)
}

func (router *Router) AddEnvironment(c *gin.Context) {
	req := &models.AddEnvironmentRequest{}

	if err := c.ShouldBind(&req); err != nil {
		writeError(c, models.NewApiError(400, "invalid arguments", err.Error()))
		return
	}

	result, apiErr := router.service.AddEnvironment(req)
	if apiErr != nil {
		writeError(c, apiErr)
		return
	}
	writeResult(c, result)
}

func (router *Router) DeleteEnvironment(c *gin.Context) {
	router.handleDelete(c, router.service.DeleteEnvironment)
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
