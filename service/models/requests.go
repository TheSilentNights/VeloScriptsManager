package models

type AddScriptRequest struct {
	Name    string `form:"name"`
	Command string `form:"command"`
	WorkDir string `form:"workdir"`
	Runner  string `form:"runner"`
}

type DeleteScriptRequest struct {
	ID string `form:"id" json:"id"`
}
