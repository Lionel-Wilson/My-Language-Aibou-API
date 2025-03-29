package render

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	headerContentType                = "Content-Type"
	contentTypeJSON                  = "application/json"
	StatusInternalServerErrorPayload = "something went wrong. please try again later."
)

func Json(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set(headerContentType, contentTypeJSON)

	body, err := json.Marshal(payload)
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = []byte(fmt.Sprintf(`{"error":"%s"}`, err))
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}
