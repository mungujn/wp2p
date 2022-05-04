package server

import (
	"net/http"
)

const (
	plainText  = "text/plain"
)

// SendResponse - sends a response
func SendResponse(w http.ResponseWriter, statusCode int, responseType string, response []byte) {
	w.Header().Set("Content-Type", responseType)
	w.WriteHeader(statusCode)
	_, _ = w.Write(response)
}
