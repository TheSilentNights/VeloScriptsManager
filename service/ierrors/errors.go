package ierrors

import (
	"net/http"

	"github/TheSilentNights/VeloScriptsManager/service/models"
)

var (
	// DbError 数据库操作失败
	DbError = models.NewApiError(http.StatusInternalServerError, "db error")
	// InvalidArgument 请求参数解析失败或不合法
	InvalidArgument = models.NewApiError(http.StatusBadRequest, "arguments are not valid")
	// IdRequired 缺少 id 参数
	IdRequired = models.NewApiError(http.StatusBadRequest, "id is required")
)
