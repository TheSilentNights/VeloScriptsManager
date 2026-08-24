package storage

type Script struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Command        string      `json:"command"`
	WorkDir        string      `json:"workDir"`
	Runner         string      `json:"runner"`
}
