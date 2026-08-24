package storage

type Script struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	WorkDir string   `json:"workDir"`
	Runner  string   `json:"runner"`
	Params  []string `json:"params"`
}

type Environment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
