package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/raymondwongso/goerp/auth"
	"github.com/raymondwongso/goerp/domain/xhttp"
)

func envDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("invalid %s %q, using default %s: %v", key, v, defaultVal, err) // #nosec G706 -- v is formatted with %q which escapes all special characters
		return defaultVal
	}
	return d
}

func main() {
	ctx := context.Background()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tracer := noop.NewTracerProvider().Tracer("")

	mux := http.NewServeMux()

	if err := auth.RegisterHTTPHandlers(ctx, mux, db, tracer, auth.Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}); err != nil {
		log.Fatalf("failed to register auth handlers: %v", err)
	}

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           xhttp.LoggingMiddleware(xhttp.CORSMiddleware(corsOrigins, mux)),
		ReadHeaderTimeout: envDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		ReadTimeout:       envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:       envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("starting server on", addr) // #nosec G706
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
