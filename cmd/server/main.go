package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"subscriptions/internal/config"
	"subscriptions/internal/database"
	"subscriptions/internal/handler"
	"subscriptions/internal/logger"
	"subscriptions/internal/repository"
	"subscriptions/internal/router"
	"subscriptions/internal/service"
)

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString("fatal: " + err.Error() + "\n")
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("starting subscriptions service", "port", cfg.HTTP.Port, "log_level", cfg.Log.Level)

	if err := database.Migrate(cfg.DB.DSN(), log); err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DB.DSN(), log)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.NewSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo, log)
	h := handler.NewSubscriptionHandler(svc, log)
	r := router.New(h, log)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case sig := <-stop:
		log.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		return err
	}

	log.Info("server stopped gracefully")
	return nil
}
