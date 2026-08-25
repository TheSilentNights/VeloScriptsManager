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
	Runner       string   `json:"runner"`
	Params       []string `json:"params"`
	Environments []string `json:"environments"` // environment ids to apply before running
}

type Environment struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Path     string   `json:"path"`
	Env      []EnvVar `json:"env"`      // key-value pairs contributed by this env
	Children []string `json:"children"` // ids of other environments to inherit from
}
