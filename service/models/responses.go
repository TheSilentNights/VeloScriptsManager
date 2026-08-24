package models

import "net/http"

// ApiError 表示服务层产生的业务错误。
// 采用统一的 {code, message, data} 信封格式，与 gin 完全解耦；
// code 同时作为 HTTP 状态码，data 可携带任意补充信息。
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (e *ApiError) Error() string {
	return e.Message
}

// NewApiError 根据状态码与消息构造业务错误，data 默认留空（nil）
func NewApiError(code int, message string) *ApiError {
	return &ApiError{Code: code, Message: message}
}

// Result 表示成功响应，采用统一的 {code, message, data} 信封格式；
// Data（any）用于保存需要返回给前端的数据。
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// NewResult 构造成功响应，code 为 200，message 默认为 "ok"
func NewResult(data any) *Result {
	return &Result{Code: http.StatusOK, Message: "ok", Data: data}
}

// NewResultWithMessage 构造带自定义 message 的成功响应
func NewResultWithMessage(message string, data any) *Result {
	return &Result{Code: http.StatusOK, Message: message, Data: data}
}