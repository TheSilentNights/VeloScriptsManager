package models

import "github/TheSilentNights/VeloScriptsManager/service/storage"

type AddScriptRequest struct {
	Name           string   `form:"name" json:"name"`
	WorkDir        string   `form:"workdir" json:"workDir"`
	Command        []string `form:"command" json:"command"`
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // environment ids to apply
}

type AddEnvironmentRequest struct {
	Name  string           `form:"name" json:"name"`
	Paths []string         `form:"paths" json:"paths"`
	Env   []storage.EnvVar `form:"env" json:"env" `
}

type UpdateEnvironmentRequest struct {
	Id    string           `form:"id" json:"id"`
	Name  string           `form:"name" json:"name"`
	Paths []string         `form:"paths" json:"paths"`
	Env   []storage.EnvVar `form:"env" json:"env" `
}

type UpdateScriptRequest struct {
	Id             string   `form:"id" json:"id"`
	Name           string   `form:"name" json:"name"`
	WorkDir        string   `form:"workdir" json:"workdir"`
	Command        []string `form:"command" json:"command"`
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // environment ids to apply
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}

type ExecuteScriptRequest struct {
	Id             string   `form:"id" json:"id"`
	Command        []string `form:"command" json:"command"`               // 可选：覆盖脚本存储的参数，传了就用传递的
	EnvironmentsId []string `form:"environmentsid" json:"environmentsid"` // 可选：覆盖脚本存储的环境 id 列表
}
