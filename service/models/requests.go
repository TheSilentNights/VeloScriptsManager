package models

type AddScriptRequest struct {
	Name    string `form:"name"`
	Command string `form:"command"`
	WorkDir string `form:"workdir"`
}

type DeleteScriptRequest struct {
	ID string `form:"id" json:"id"`
}
