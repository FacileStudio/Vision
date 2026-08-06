package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/database"
	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/apikeys"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/Vision/apps/api/modules/webhooks"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/Vision/apps/api/schemas"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
)

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}

	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return
	}

	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	if err := os.MkdirAll(filepath.Join(appEnv.StorageDir, "avatars"), 0o755); err != nil {
		appLogger.Error("failed to create avatar directory", slog.Any("error", err))
		return
	}

	authService := auth.NewService(db, appEnv.StorageDir, appLogger)
	workspaceService := workspaces.NewService(db)
	siteService := sites.NewService(db)
	eventHub := events.NewHub()
	activeTracker := events.NewActiveTracker()
	eventService := events.NewService(db)
	eventService.SetHub(eventHub)
	analyticsService := analytics.NewService(db)
	goalService := goals.NewService(db)
	apiKeyService := apikeys.NewService(db)
	middleware.SetAPIKeyAuthenticator(apiKeyService)
	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: []string{appEnv.Domain},
		},
	})

	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, referenceConfig())

	avatarFS := http.StripPrefix("/files/", http.FileServer(http.Dir(appEnv.StorageDir)))
	router.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		avatarFS.ServeHTTP(w, r)
	})

	auth.RegisterRoutes(router, authService, appEnv)
	workspaces.RegisterRoutes(router, workspaceService, authService)
	sites.RegisterRoutes(router, siteService, authService)
	events.RegisterRoutes(router, eventService, eventHub, activeTracker, authService, db)
	analytics.RegisterRoutes(router, analyticsService, activeTracker, db, authService)

	goals.RegisterRoutes(router, goalService, authService)
	apikeys.RegisterRoutes(router, apiKeyService, authService)

	webhookService := webhooks.NewService(db)
	webhooks.RegisterRoutes(router, webhookService, analyticsService, authService)
	webhooks.StartScheduler(db, analyticsService)

	addr := ":" + appEnv.Port
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return
		}
		appLogger.Info("server stopped")
	}
}

// referenceConfig describes the API reference served at /docs. Registry paths
// are written relative to /api, the one server every documented route hangs off.
func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Vision API",
		Description: "Self-hosted, privacy-friendly web analytics.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{Modules: []documentation.Module{
			auth.Documentation,
			sites.Documentation,
			events.Documentation,
			analytics.Documentation,
			goals.Documentation,
			apikeys.Documentation,
		}},
	}
}
