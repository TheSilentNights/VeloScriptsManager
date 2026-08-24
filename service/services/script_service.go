package services

import (
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
	"github/TheSilentNights/VeloScriptsManager/service/utils"

	"github.com/gin-gonic/gin"
)

func (service *Service) ListScripts(c *gin.Context) {
	list, err := service.scriptRepo.List()

	if err != nil {
		c.Errors.JSON()
	}

	c.JSON(200, list)
}

func (service *Service) AddScript(req *models.AddScriptRequest) error {

	script := storage.Script{
		ID:      utils.GenerateScriptId(),
		Name:    req.Name,
		Command: req.Command,
		WorkDir: req.WorkDir,
	}

	if err := service.scriptRepo.Upsert(script); err != nil {
		return errors.New(ierrors.DbError)
	}

	return nil
}
