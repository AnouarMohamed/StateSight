package apihttp

import (
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *errorBody     `json:"error,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeSuccess(w http.ResponseWriter, status int, data any, meta map[string]any) {
	writeJSON(w, status, apiResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func writeError(w http.ResponseWriter, status int, code, message string, meta map[string]any) {
	writeJSON(w, status, apiResponse{
		Success: false,
		Error: &errorBody{
			Code:    code,
			Message: message,
		},
		Meta: meta,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
