package response

import (
	"encoding/json"
	"net/http"

	"github.com/erewhile/iam/pkg/response/code"
)

type response[T any] struct {
	Code    code.Code `json:"code"`
	Status  bool      `json:"status"`
	Message string    `json:"message"`
	Data    T         `json:"data,omitempty"`
}

func write[T any](w http.ResponseWriter, statusCode int, resp response[T]) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

func OK(w http.ResponseWriter) {
	write(w, http.StatusOK, response[any]{
		Code:    code.Success,
		Status:  true,
		Message: code.Success.Message(),
		Data:    nil,
	})
}

func OkData[T any](w http.ResponseWriter, data T) {
	write(w, http.StatusOK, response[T]{
		Code:    code.Success,
		Status:  true,
		Message: code.Success.Message(),
		Data:    data,
	})
}

func Fail(w http.ResponseWriter, errCode code.Code) {
	write(w, http.StatusOK, response[any]{
		Code:    errCode,
		Status:  false,
		Message: errCode.Message(),
		Data:    nil,
	})
}

func FailMessage(w http.ResponseWriter, errCode code.Code, message string) {
	write(w, http.StatusOK, response[any]{
		Code:    errCode,
		Status:  false,
		Message: message,
		Data:    nil,
	})
}

func Custom(w http.ResponseWriter, httpStatus int, message string) {
	write(w, httpStatus, response[any]{
		Code:    code.Custom,
		Status:  false,
		Message: message,
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	write(w, http.StatusBadRequest, response[any]{
		Code:    code.BadRequest,
		Status:  false,
		Message: message,
		Data:    nil,
	})
}

func Unauthorized(w http.ResponseWriter) {
	write(w, http.StatusUnauthorized, response[any]{
		Code:    code.Unauthorized,
		Status:  false,
		Message: code.Unauthorized.Message(),
		Data:    nil,
	})
}

func Forbidden(w http.ResponseWriter) {
	write(w, http.StatusForbidden, response[any]{
		Code:    code.Forbidden,
		Status:  false,
		Message: code.Forbidden.Message(),
		Data:    nil,
	})
}

func NotFound(w http.ResponseWriter) {
	write(w, http.StatusNotFound, response[any]{
		Code:    code.NotFound,
		Status:  false,
		Message: code.NotFound.Message(),
		Data:    nil,
	})
}

func InternalServer(w http.ResponseWriter) {
	write(w, http.StatusInternalServerError, response[any]{
		Code:    code.InternalServerError,
		Status:  false,
		Message: code.InternalServerError.Message(),
		Data:    nil,
	})
}
