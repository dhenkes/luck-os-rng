package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Code.HTTPStatus(), appErr)
		return
	}
	var valErr *model.ValidationErrors
	if errors.As(err, &valErr) {
		writeJSON(w, http.StatusBadRequest, model.AppError{
			Code:    model.ErrorCodeInvalidArgument,
			Message: valErr.Error(),
			Details: valErr.Fields(),
		})
		return
	}
	slog.Error("internal error", "error", err)
	writeJSON(w, http.StatusInternalServerError, model.AppError{
		Code:    model.ErrorCodeInternal,
		Message: "internal server error",
	})
}

func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func isSSE(r *http.Request) bool {
	return r.URL.Query().Get("sse") == "1"
}

// streamOrPage either streams ANSI (curl), SSE (browser JS), or serves the browser page.
func streamOrPage(w http.ResponseWriter, r *http.Request, title, path, configForm string, frames []renderer.Frame) {
	if isSSE(r) {
		renderer.StreamSSE(w, frames)
		return
	}
	if wantsHTML(r) {
		renderer.BrowserPage(w, title, path, configForm)
		return
	}
	renderer.StreamFrames(w, frames)
}
