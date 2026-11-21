// Package httpapi is a fixture of the error vocabulary carriers.
package httpapi

import "net/http"

type APIError struct {
	Status int
	Code   string
}

func NotFound(code string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: code}
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any) error {
	return nil
}

func WriteAPIError(w http.ResponseWriter, r *http.Request, apiErr *APIError) {}
