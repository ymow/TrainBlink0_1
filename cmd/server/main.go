package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ymow/messenger_protocol_research/internal/api"
	"github.com/ymow/messenger_protocol_research/internal/cache"
	"github.com/ymow/messenger_protocol_research/internal/config"
	"github.com/ymow/messenger_protocol_research/internal/database"
	"github.com/ymow/messenger_protocol_research/pkg/logger"
)

const (
	version = "0.1.0"
	banner  = `
╔══════════════════════════════════════════════════════╗
║                                                      ║
║     🚄 TrainBlink Server - Phase 0: Hello World     ║
║                                                      ║
║     Version: %s                                ║
║     Stage:   HTTP + WebSocket Foundation            ║
║                                                      ║
╚══════════════════════════════════════════════════════╝
`
)

func main() {
	// Print banner
	fmt.Printf(banner, version)

	// Initialize logger
	if err := logger.Initialize("info"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting TrainBlink Server",
		zap.String("version", version),
		zap.String("stage", "Phase 0 - Hello World"),
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize PostgreSQL
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := database.Close(db); err != nil {
			logger.Error("Failed to close database", zap.Error(err))
		}
	}()

	// Initialize Redis
	redisService, err := cache.NewRedisService(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer func() {
		if err := redisService.Close(); err != nil {
			logger.Error("Failed to close Redis", zap.Error(err))
		}
	}()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes with database and Redis
	api.SetupRoutes(router, db, redisService.GetClient())

	// Create HTTP server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server is running",
			zap.String("address", server.Addr),
			zap.String("mode", gin.Mode()),
		)
		logger.Info("Try these endpoints:")
		logger.Info("  • GET  http://localhost:8080/ping")
		logger.Info("  • GET  http://localhost:8080/health")
		logger.Info("  • GET  http://localhost:8080/api/v1/hello")
		logger.Info("  • GET  http://localhost:8080/api/v1/welcome?name=Alice")
		logger.Info("  • POST http://localhost:8080/api/v1/trips/start")
		logger.Info("  • GET  http://localhost:8080/api/v1/trips/active")
		logger.Info("  • GET  http://localhost:8080/api/v1/discoveries/stats")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited successfully")
}
