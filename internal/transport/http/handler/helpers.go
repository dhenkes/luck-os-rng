package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dhenkes/luck-os-rng/internal/model"
	"github.com/dhenkes/luck-os-rng/internal/renderer"
)

// configuredHost is the trusted host value set at startup via SetHost.
// All URL generation uses this instead of reading r.Host from the request.
var configuredHost = "localhost"

// SetHost sets the trusted host used for generating URLs.
func SetHost(host string) {
	configuredHost = host
}

// getHost returns the configured trusted host.
func getHost() string {
	return configuredHost
}

func parseEngagementState(r *http.Request) model.EngagementState {
	return model.ParseEngagementState(r.URL.Query())
}

// buildNextURL constructs a "play again" URL with game config + engagement state.
func buildNextURL(host, basePath string, gameParams string, eng model.EngagementState) string {
	engQS := eng.QueryString()
	var qs string
	switch {
	case gameParams != "" && engQS != "":
		qs = gameParams + "&" + engQS
	case gameParams != "":
		qs = gameParams
	case engQS != "":
		qs = engQS
	}
	if qs != "" {
		return host + basePath + "?" + qs
	}
	return host + basePath
}

// buildDoubleURL constructs a double-or-nothing URL.
func buildDoubleURL(host string, stake int, eng model.EngagementState) string {
	u := fmt.Sprintf("%s/double?stake=%d", host, stake)
	engQS := eng.QueryString()
	if engQS != "" {
		u += "&" + engQS
	}
	return u
}

// gameQueryString extracts game-specific query params (excluding engagement params).
func gameQueryString(r *http.Request, engKeys ...string) string {
	skip := map[string]bool{"sse": true, "s": true, "sc": true, "h": true, "u": true, "bet": true}
	for _, k := range engKeys {
		skip[k] = true
	}
	var parts []string
	for k, vs := range r.URL.Query() {
		if skip[k] || len(vs) == 0 || vs[0] == "" {
			continue
		}
		parts = append(parts, k+"="+vs[0])
	}
	return strings.Join(parts, "&")
}

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

func isFast(r *http.Request) bool {
	return r.URL.Query().Get("fast") == "1"
}

func isInstant(r *http.Request) bool {
	return r.URL.Query().Get("instant") == "1"
}

// streamOrPage either streams ANSI (curl), SSE (browser JS), or serves the browser page.
func streamOrPage(w http.ResponseWriter, r *http.Request, title, path, configForm string, frames []renderer.Frame) {
	if isSSE(r) {
		renderer.StreamSSE(w, frames, isFast(r), isInstant(r))
		return
	}
	if wantsHTML(r) {
		renderer.BrowserPage(w, title, path, configForm)
		return
	}
	renderer.StreamFrames(w, frames, isFast(r), isInstant(r))
}
