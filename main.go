// Command golang-tut is the entry point for the Go tutorial app.
//
// This file does ONE job: composition root. It loads configuration,
// constructs each domain object, and starts the HTTP server. All real
// logic lives in internal/* packages.
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-tut/internal/api"
	"golang-tut/internal/config"
	"golang-tut/internal/lesson"
	"golang-tut/internal/runner"
	"golang-tut/internal/tutor"
)

//go:embed web
var webFS embed.FS

func main() {
	addrFlag := flag.String("addr", "", "listen address (default: :$PORT or :8080)")
	flag.Parse()

	// 1. Configuration: optional .env file, real env vars always win.
	if n, err := config.LoadDotEnv(".env"); err != nil {
		log.Printf("warning: reading .env: %v", err)
	} else if n > 0 {
		log.Printf("loaded %d entries from .env", n)
	}

	// 2. Resolve listen address. CLI flag > $PORT (Cloud Run) > :8080.
	addr := resolveAddr(*addrFlag)

	// 3. Domain objects.
	lessons := lesson.NewInMemoryRepository(lesson.Catalog)
	codeRunner := runner.NewFromEnv(10 * time.Second)
	chat := tutor.NewService(tutor.SelectFromEnv(), lessons)

	log.Printf("code runner: %s", codeRunner.Backend())
	if s := chat.Status(); s.Available {
		log.Printf("AI chat enabled via %s (%s)", s.Provider, s.Model)
	} else {
		log.Printf("AI chat disabled: %s", s.Hint)
	}

	// 4. Transport (HTTP): handed the domain objects it needs.
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}
	server := api.NewServer(lessons, codeRunner, chat, staticFS)

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 5. Serve in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("Go tutorial listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// 6. Graceful shutdown. Cloud Run / Kubernetes send SIGTERM ~10s before
	// killing the container; we drain in-flight requests before exiting.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutdown signal received, draining...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("bye")
}

// resolveAddr picks the listen address with this precedence:
//
//  1. CLI flag (-addr) if provided
//  2. $PORT environment variable (Cloud Run, App Engine, Heroku, etc.)
//  3. :8080 fallback
func resolveAddr(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}
