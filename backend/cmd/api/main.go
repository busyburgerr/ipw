// Command api is the HTTP entrypoint for the freelance-marketplace backend.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ipw/internal/auth"
	"ipw/internal/catalog"
	"ipw/internal/config"
	"ipw/internal/httpx"
	"ipw/internal/platform/postgres"
	"ipw/internal/platform/redis"
	"ipw/internal/platform/storage"
	"ipw/internal/profile"
	"ipw/internal/project"
	"ipw/internal/user"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("fatal", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log.Info("config loaded", slog.String("env", cfg.Env))

	db, err := postgres.Open(cfg.DB, cfg.Env)
	if err != nil {
		return err
	}
	log.Info("postgres connected")

	if err := postgres.Migrate(db); err != nil {
		return err
	}
	log.Info("migrations applied")

	rdb, err := redis.Open(cfg.Redis)
	if err != nil {
		return err
	}
	log.Info("redis connected")
	defer func() { _ = rdb.Close() }()

	ctx := context.Background()
	files, err := storage.New(ctx, cfg.Storage)
	if err != nil {
		return err
	}
	log.Info("object storage ready")

	app := httpx.NewServer(cfg, log)

	// --- Feature wiring -------------------------------------------------------
	users := user.NewPostgresStore(db)

	authSvc := auth.NewService(users, db, cfg.Auth)
	authMW := auth.NewMiddleware(cfg.Auth)
	auth.NewHandler(authSvc, authMW, cfg.Auth).Register(app)

	catalogStore := catalog.NewPostgresStore(db)
	if err := catalog.Seed(ctx, catalogStore); err != nil {
		return err
	}
	log.Info("catalog seeded")
	catalog.NewHandler(catalogStore).Register(app)

	profileStore := profile.NewPostgresStore(db)
	profileSvc := profile.NewService(profileStore, catalogStore)
	profile.NewHandler(profileSvc, users, catalogStore, files, authMW).Register(app)

	projectStore := project.NewPostgresStore(db)
	projectSvc := project.NewService(projectStore, catalogStore)
	project.NewHandler(projectSvc, authMW).Register(app)
	// Additional feature routers are registered here as they are built.
	// -----------------------------------------------------------------------

	serverErr := make(chan error, 1)
	go func() { serverErr <- app.Listen(":" + cfg.HTTP.Port) }()
	log.Info("listening", slog.String("port", cfg.HTTP.Port))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-stop:
		log.Info("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		log.Info("shutdown complete")
		return nil
	}
}
