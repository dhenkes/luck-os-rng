package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSlotsHandler(t *testing.T) {
	r := chi.NewRouter()
	NewSlotsHandler().Register(r)

	req := httptest.NewRequest("GET", "/slots?instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "GO AGAIN") {
		t.Fatal("response should contain GO AGAIN URL")
	}
	if !strings.Contains(body, "Score:") {
		t.Fatal("response should contain score display")
	}
}

func TestSlotsHandlerWithEngagement(t *testing.T) {
	r := chi.NewRouter()
	NewSlotsHandler().Register(r)

	req := httptest.NewRequest("GET", "/slots?s=2&sc=500&h=WW&bet=high&instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "HIGH") {
		t.Fatal("response should show bet tier")
	}
}

func TestSlotsHandlerHighLuckNoUnlockRequired(t *testing.T) {
	r := chi.NewRouter()
	NewSlotsHandler().Register(r)

	// High luck should work without any unlock params.
	req := httptest.NewRequest("GET", "/slots?luck=high&instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	// Should contain game results, not a lock screen.
	// The old behavior would show "LOCKED: HIGH LUCK" but now luck modes are always available.
	// Should contain game results (grid, score, etc.)
	if !strings.Contains(body, "Score:") {
		t.Fatal("should show game results with engagement data")
	}
}

func TestRouletteHandler(t *testing.T) {
	r := chi.NewRouter()
	NewRouletteHandler().Register(r)

	req := httptest.NewRequest("GET", "/roulette?instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "LUCK") {
		t.Fatal("response should contain LUCK header")
	}
}

func TestCoinFlipHandler(t *testing.T) {
	r := chi.NewRouter()
	NewCoinFlipHandler().Register(r)

	req := httptest.NewRequest("GET", "/coinflip?instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestDiceHandler(t *testing.T) {
	r := chi.NewRouter()
	NewDiceHandler().Register(r)

	req := httptest.NewRequest("GET", "/dice?instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestDoubleHandler(t *testing.T) {
	r := chi.NewRouter()
	NewDoubleHandler().Register(r)

	req := httptest.NewRequest("GET", "/double?stake=300&sc=500&s=2&h=WW&instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	// Should contain either WIN or LOST.
	if !strings.Contains(body, "YOU WON") && !strings.Contains(body, "YOU LOST") {
		t.Fatal("double handler should show win or loss result")
	}
}

func TestDoubleHandlerDefaultStake(t *testing.T) {
	r := chi.NewRouter()
	NewDoubleHandler().Register(r)

	req := httptest.NewRequest("GET", "/double?instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}


func TestLandingHandlerText(t *testing.T) {
	h := NewLandingHandler()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "LUCK") {
		t.Fatal("landing should contain LUCK header")
	}
	if !strings.Contains(body, "DOUBLE") {
		t.Fatal("landing should mention double or nothing")
	}
}

func TestLandingHandlerHTML(t *testing.T) {
	h := NewLandingHandler()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Fatal("HTML landing should contain <html> tag")
	}
	if !strings.Contains(body, "/double") {
		t.Fatal("HTML landing should link to double")
	}
}

func TestBrowserPageSSE(t *testing.T) {
	r := chi.NewRouter()
	NewSlotsHandler().Register(r)

	req := httptest.NewRequest("GET", "/slots?sse=1&instant=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestBrowserPageHTML(t *testing.T) {
	r := chi.NewRouter()
	NewSlotsHandler().Register(r)

	req := httptest.NewRequest("GET", "/slots", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Fatal("HTML page should contain <html> tag")
	}
	if !strings.Contains(body, "bet") {
		t.Fatal("HTML form should contain bet selector")
	}
}

func TestGameQueryString(t *testing.T) {
	req := httptest.NewRequest("GET", "/slots?mode=standard&rows=3&cols=5&bet=high&s=2&sc=500&sse=1&h=WW&u=high", nil)
	qs := gameQueryString(req)

	// Should include game params but not engagement/sse params.
	if strings.Contains(qs, "bet=") {
		t.Fatal("gameQueryString should exclude bet")
	}
	if strings.Contains(qs, "sse=") {
		t.Fatal("gameQueryString should exclude sse")
	}
	if strings.Contains(qs, "sc=") {
		t.Fatal("gameQueryString should exclude sc")
	}
	if strings.Contains(qs, "h=") {
		t.Fatal("gameQueryString should exclude h")
	}
	if strings.Contains(qs, "u=") {
		t.Fatal("gameQueryString should exclude u")
	}
	// s= is tricky because it matches "rows=3" etc. Check for exact "s=2" instead.
	for _, part := range strings.Split(qs, "&") {
		if part == "s=2" {
			t.Fatal("gameQueryString should exclude s engagement param")
		}
	}
}
