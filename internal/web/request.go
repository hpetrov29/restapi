package web

import (
	"net/http"

	"github.com/go-chi/chi"
)

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}