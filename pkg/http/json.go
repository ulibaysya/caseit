package http

import (
	"net/http"

	pkgjson "github.com/ulibaysya/caseit/pkg/json"
)

func ErrorJSON(w http.ResponseWriter, error string, code int) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	pkgjson.Write(w, error) //nolint:errcheck
}

func ErrorJSON500(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
