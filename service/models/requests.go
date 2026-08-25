package models

import "github/TheSilentNights/VeloScriptsManager/service/storage"

type AddScriptRequest struct {
	Name         string   `form:"name"`
	WorkDir      string   `form:"workdir"`
	Runner       string   `form:"runner"`
	Params       []string `form:"params" json:"params"`
	Environments []string `json:"environments"` // environment ids to apply
}

type AddEnvironmentRequest struct {
	Name     string           `form:"name"`
	Type     string           `form:"type"`
	Path     string           `form:"path"`
	Env      []storage.EnvVar `json:"env"`
	Children []string         `json:"children"`
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}

type ExecuteScriptRequest struct {
	Id string `form:"id" json:"id"`
}
