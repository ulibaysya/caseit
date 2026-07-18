package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON[T any](w http.ResponseWriter, v T) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding data to json: %w", err)
	}
	if n, err := w.Write(encoded); err != nil {
		return fmt.Errorf("writing json data: %w", err)
	} else if encodedLen := len(encoded); n != encodedLen {
		return fmt.Errorf("written %v bytes out of %v", n, encodedLen)
	}
	return nil
}

func ErrorJSON(w http.ResponseWriter, error string, code int) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	encodedError, _ := json.Marshal(error)
	w.Write(encodedError) //nolint:errcheck
}

func ErrorJSON500(w http.ResponseWriter) {
	ErrorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
