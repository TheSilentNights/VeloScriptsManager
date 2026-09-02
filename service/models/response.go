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
