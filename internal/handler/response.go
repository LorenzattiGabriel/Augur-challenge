package handler

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeInternalServer = "INTERNAL_ERROR"
	ErrCodeRateLimited    = "RATE_LIMITED"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondSuccess(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

func respondNotFound(w http.ResponseWriter, message string) {
	respondError(w, http.StatusNotFound, ErrCodeNotFound, message)
}

func respondBadRequest(w http.ResponseWriter, message string) {
	respondError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
}

func respondInternalError(w http.ResponseWriter) {
	respondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Internal server error")
}
