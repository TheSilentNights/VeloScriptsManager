package storage

type Script struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Command        string      `json:"command"`
	WorkDir        string      `json:"workDir"`
	Runner         string      `json:"runner"`
}

type Environment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
