package models

import "github/TheSilentNights/VeloScriptsManager/service/storage"

type AddScriptRequest struct {
	Name         string   `form:"name" json:"name"`
	WorkDir      string   `form:"workdir" json:"workdir"`
	Runner       string   `form:"runner" json:"runner"`
	Params       []string `form:"params" json:"params"`
	Environments []string `form:"environments" json:"environments"` // environment ids to apply
}

type AddEnvironmentRequest struct {
	Name     string           `form:"name" json:"name"`
	Type     string           `form:"type" json:"type"`
	Path     string           `form:"path" json:"path"`
	Env      []storage.EnvVar `form:"env" json:"env" `
	Children []string         `form:"children" json:"children"`
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}

type ExecuteScriptRequest struct {
	Id string `form:"id" json:"id"`
}
