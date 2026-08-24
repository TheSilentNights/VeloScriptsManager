package models

type AddScriptRequest struct {
	Name    string `form:"name"`
	Command string `form:"command"`
	WorkDir string `form:"workdir"`
}
