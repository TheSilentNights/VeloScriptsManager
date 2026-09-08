package models

import "time"

type ExecutionStatusInfo struct {
	ExecutionId  string    `json:"executionId"`
	ScriptId     string    `json:"scriptId"`
	Name         string    `json:"name"`
	StartedAt    time.Time `json:"startedAt"`
	Command      []string  `json:"command"`
	Environments []string  `json:"environments"`
	Status       string    `json:"status"`   // running | finished | failed
	ExitCode     int       `json:"exitCode"` // -1 while still running
	Error        string    `json:"error"`
}

type EventInfoResponse struct {
	EventId   string      `json:"eventId"`
	Type      string      `json:"type"`
	EventData interface{} `json:"event"`
}

type FileChangeEventInfo struct {
	Path string `json:"path"`
}

type TimeEventInfo struct {
	Interval time.Duration `json:"interval"`
	Repeat   bool          `json:"repeat"`
}
