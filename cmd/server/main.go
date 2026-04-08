package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/dhenkes/luck-os-rng/internal/transport/http/handler"
	"github.com/dhenkes/luck-os-rng/internal/transport/http/middleware"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("luck starting")

	landingHandler := handler.NewLandingHandler()
	rouletteHandler := handler.NewRouletteHandler()
	slotsHandler := handler.NewSlotsHandler()
	coinflipHandler := handler.NewCoinFlipHandler()
	diceHandler := handler.NewDiceHandler()

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Landing page.
	r.Get("/", landingHandler.ServeHTTP)

	// Game routes.
	rouletteHandler.Register(r)
	slotsHandler.Register(r)
	coinflipHandler.Register(r)
	diceHandler.Register(r)

	srv := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("listening", "addr", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("stopped")
}
