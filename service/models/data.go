package models

type ScriptOption struct {
	ID         string   `json:"id"`
	ScriptName string   `json:"scriptName"`
	WorkDir    string   `json:"workDir"`
	Command    []string `json:"command"`
	Env        []string `json:"env"`
}
