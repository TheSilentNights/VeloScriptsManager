package models

type AddScriptRequest struct {
	Name    string   `form:"name"`
	WorkDir string   `form:"workdir"`
	Runner  string   `form:"runner"`
	Params  []string `form:"params" json:"params"`
}

type AddEnvironmentRequest struct {
	Name string `form:"name"`
	Type string `form:"type"`
	Path string `form:"path"`
}

type DeleteRequest struct {
	Id string `form:"id" json:"id"`
}

type ExecuteScriptRequest struct {
	Id string `form:"id" json:"id"`
}
