package models

// ApiError 表示服务层产生的业务错误。
// 它同时携带返回给前端的 HTTP 状态码与消息，与 gin 完全解耦；
// router 层只需原样写入响应即可，无需自行判断错误类型。
type ApiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	return e.Message
}

// NewApiError 根据状态码与消息构造业务错误
func NewApiError(status int, message string) *ApiError {
	return &ApiError{Status: status, Message: message}
}