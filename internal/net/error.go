package net

import (
	"encoding/json"
	"net/http"
)

type StatusCode = int

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, code StatusCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
	})
}