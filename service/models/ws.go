package models

// WsServerFrame is sent from the server to the client on the execution attach
// WebSocket. Type is "output" while the process is running or "exit" once it
// has finished. Output chunks are base64-encoded raw bytes so that non-UTF-8
// (e.g. GBK console) output survives the JSON round trip intact.
type WsServerFrame struct {
	Type  string `json:"type"`
	Data  string `json:"data,omitempty"`
	Code  int    `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
}

// WsClientFrame is sent from the client to the server on the execution attach
// WebSocket. Type is one of:
//   - "stdin":      Data holds base64-encoded bytes written to the process stdin
//   - "close_stdin": close the process stdin (EOF for interactive programs)
//   - "kill":       force-terminate the process
type WsClientFrame struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}
