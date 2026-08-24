package utils

import "github.com/google/uuid"

func GenerateScriptId() string {
	return uuid.NewString()
}

func GenerateEnvironmentId() string {
	return uuid.NewString()
}
