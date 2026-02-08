package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopherpass/internal/config"
	"gopherpass/internal/repository/postgres"
	"gopherpass/internal/security"
	"gopherpass/internal/service"
	httpTransport "gopherpass/internal/transport/http"
	"gopherpass/internal/transport/http/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Database
	pool, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		log.Fatal("db connection failed:", err)
	}
	defer pool.Close()

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("db ping failed:", err)
	}
	log.Println("database connected")

	// Security
	jwtMgr := security.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)
	crypto, err := security.NewCrypto([]byte(cfg.DataEncryptionKey))
	if err != nil {
		log.Fatal("crypto init failed:", err)
	}

	// Repositories
	userRepo := postgres.NewUserRepo(pool)
	secretRepo := postgres.NewSecretRepo(pool)

	// Services
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	dataSvc := service.NewDataService(secretRepo, crypto)
	syncSvc := service.NewSyncService(secretRepo, crypto)

	// Handlers
	authH := handlers.NewAuthHandler(authSvc)
	dataH := handlers.NewDataHandler(dataSvc)
	syncH := handlers.NewSyncHandler(syncSvc)

	// Router
	router := httpTransport.NewRouter(authH, dataH, syncH, jwtMgr)

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("shutdown error:", err)
	}
	log.Println("server stopped")
}
