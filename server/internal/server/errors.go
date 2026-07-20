package server

import (
	"net/http"

	"github.com/kindling/kindling/pkg/types"
)

func WriteError(w http.ResponseWriter, code string, message string) {
	status := ErrorHTTPStatus(code)
	WriteJSON(w, status, types.ErrorResponse{
		Success: false,
		Code:    code,
		Error:   message,
	})
}

func ErrorHTTPStatus(code string) int {
	switch code {
	case types.ErrCollectionRequired,
		types.ErrCollectionProjectMismatch,
		types.ErrInvalidContentType:
		return http.StatusBadRequest

	case types.ErrAuthInvalidToken,
		types.ErrAuthSessionInvalid:
		return http.StatusUnauthorized

	case types.ErrPayloadTooLarge:
		return http.StatusRequestEntityTooLarge

	case types.ErrFirestoreQuota:
		return http.StatusTooManyRequests

	case types.ErrParseFailed,
		types.ErrEmptyJSON:
		return http.StatusBadRequest

	case types.ErrInternal:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
