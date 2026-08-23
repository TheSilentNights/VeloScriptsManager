package services

import "github/TheSilentNights/VeloScriptsManager/service/storage"

type Service struct {
	scriptRepo *storage.ScriptRepo
}

func NewService(repo *storage.ScriptRepo) *Service {
	return &Service{repo}
}
