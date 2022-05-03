package server

import (
	"net/http"
)

const (
	contentTypeJSON = "application/json;charset=utf-8"
)

func SendEmptyResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
}

// SendRawResponse - common method for writing any raw ([]byte) response.
func SendRawResponse(w http.ResponseWriter, statusCode int, binBody []byte) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	_, _ = w.Write(binBody)
}
