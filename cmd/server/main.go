package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/A4AD-team/profile-service/internal/config"
	"github.com/A4AD-team/profile-service/internal/handler"
	"github.com/A4AD-team/profile-service/internal/messaging"
	"github.com/A4AD-team/profile-service/internal/repository"
	"github.com/A4AD-team/profile-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Logger
	var logger *zap.Logger
	if cfg.App.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Database
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		logger.Fatal("failed to parse database url", zap.Error(err))
	}
	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal("database ping failed", zap.Error(err))
	}
	logger.Info("database connected")

	// Layers
	repo := repository.NewPostgresRepo(pool)
	svc := service.NewProfileService(repo)
	profileHandler := handler.NewProfileHandler(svc)
	healthHandler := handler.NewHealthHandler(pool)

	// Router
	r := chi.NewRouter()
	handler.RegisterRoutes(r, cfg, profileHandler, healthHandler)

	// RabbitMQ
	consumer := messaging.NewConsumer(cfg.RabbitMQ.URL, svc, logger)
	if err := consumer.Connect(); err != nil {
		logger.Error("failed to connect to rabbitmq", zap.Error(err))
	} else {
		logger.Info("rabbitmq connected")
		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		defer consumerCancel()

		go func() {
			if err := consumer.Start(consumerCtx); err != nil {
				logger.Error("rabbitmq consumer stopped", zap.Error(err))
			}
		}()
	}
	defer consumer.Close()

	// HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting profile-service", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("server stopped")
}
