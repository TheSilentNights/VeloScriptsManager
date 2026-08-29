package storage

// EnvVar is a single environment variable pair applied to a process.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Script struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	WorkDir      string   `json:"workDir"`
	Command      []string `json:"command"`
	Environments []string `json:"environments"` // environment ids to apply before running
}

type Environment struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Paths []string `json:"paths"` // directories prepended to PATH when this env is applied
	Env   []EnvVar `json:"env"`   // key-value pairs contributed by this env
}
