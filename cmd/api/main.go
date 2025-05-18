package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/lib/pq"

	"github.com/dmksnnk/octo/docs"
	"github.com/dmksnnk/octo/internal/api"
	"github.com/dmksnnk/octo/internal/auth"
	"github.com/dmksnnk/octo/internal/platform/httpplatform"
	"github.com/dmksnnk/octo/internal/service"
	"github.com/dmksnnk/octo/internal/storage"
)

type config struct {
	ListenAddress string     `env:"LISTEN_ADDRESS" envDefault:":8080"`
	LogLevel      slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
	DatabaseURL   string     `env:"DATABASE_URL"`
}

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/-/health", health(db))
	mux.Handle("/docs",
		httpplatform.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFileFS(w, r, docs.Swagger, "swagger.html")
		}),
			httpplatform.LogRequests(logger),
		),
	)
	mux.Handle("/docs/openapi.yaml",
		httpplatform.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFileFS(w, r, docs.Swagger, "openapi.yaml")
		}),
			httpplatform.LogRequests(logger),
		),
	)

	pg := storage.NewPostgres(db)
	svc := service.NewService(pg)
	a := api.NewAPI(svc)
	mux.Handle("/",
		httpplatform.Wrap(
			api.NewRouter(a),
			auth.CheckUser(pg),
			httpplatform.LogRequests(logger),
		),
	)

	srv := http.Server{
		Addr:    cfg.ListenAddress,
		Handler: mux,
	}

	go func() {
		slog.Info("starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	<-rootCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
	}

	slog.Info("server shutdown")
}

func parseConfig() config {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("parse config", "error", err)
		os.Exit(1)
	}

	return cfg
}

func health(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
