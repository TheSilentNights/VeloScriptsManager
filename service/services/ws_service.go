package services

import (
	"encoding/base64"
	"sync"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/models"

	"github.com/gorilla/websocket"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

type WsStore struct {
	wsMu    sync.Mutex
	wsConns map[*websocket.Conn]struct{}
}

// StreamExecution streamExecution bridges an upgraded WebSocket to the stdio of the given
// execution: server frames carry base64 output chunks plus a final exit frame,
// client frames forward stdin/close/kill to the process.
//
// The connection is registered so CloseWebSockets can force-close it during
// shutdown, and it is always closed on return. The server pings every
// wsPingPeriod; a client that misses wsPongWait without pong is disconnected.
func (service *Service) StreamExecution(conn *websocket.Conn, execution *models.Execution) {
	service.wsService.wsMu.Lock()
	service.wsService.wsConns[conn] = struct{}{}
	service.wsService.wsMu.Unlock()

	defer func() {
		service.wsService.wsMu.Lock()
		delete(service.wsService.wsConns, conn)
		service.wsService.wsMu.Unlock()
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

// CloseWebSockets force-closes every attached execution WebSocket so a graceful
// shutdown is not blocked by long-lived connections.
func (service *Service) CloseWebSockets() {
	service.wsService.wsMu.Lock()
	defer service.wsService.wsMu.Unlock()
	for conn := range service.wsService.wsConns {
		_ = conn.Close()
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
