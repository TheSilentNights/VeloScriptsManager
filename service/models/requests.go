package models

type AddScriptRequest struct {
	Name    string `form:"name"`
	Command string `form:"command"`
	WorkDir string `form:"workdir"`
	Runner  string `form:"runner"`
}

type AddEnvironmentRequest struct {
	Name string `form:"name"`
	Type string `form:"type"`
	Path string `form:"path"`
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}
