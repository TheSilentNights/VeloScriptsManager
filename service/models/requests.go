package models

import "github/TheSilentNights/VeloScriptsManager/service/storage"

type AddScriptRequest struct {
	Name           string   `form:"name" json:"name"`
	WorkDir        string   `form:"workdir" json:"workdir"`
	Runner         string   `form:"runner" json:"runner"`
	Params         []string `form:"params" json:"params"`
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // environment ids to apply
}

type AddEnvironmentRequest struct {
	Name     string           `form:"name" json:"name"`
	Type     string           `form:"type" json:"type"`
	Path     string           `form:"path" json:"path"`
	Env      []storage.EnvVar `form:"env" json:"env" `
	Children []string         `form:"children" json:"children"`
}

type UpdateScriptRequest struct {
	Id             string   `form:"id" json:"id"`
	Name           string   `form:"name" json:"name"`
	WorkDir        string   `form:"workdir" json:"workdir"`
	Runner         string   `form:"runner" json:"runner"`
	Params         []string `form:"params" json:"params"`
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // environment ids to apply
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}

type ExecuteScriptRequest struct {
	Id             string   `form:"id" json:"id"`
	Params         []string `form:"params" json:"params"`                 // 可选：覆盖脚本存储的参数，传了就用传递的
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // 可选：覆盖脚本存储的环境 id 列表
}
