package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marathozin/notes-api-go/internal/config"
	"github.com/marathozin/notes-api-go/internal/handler"
	"github.com/marathozin/notes-api-go/internal/service"
	"github.com/marathozin/notes-api-go/internal/store/postgres"
)

// @title Notes API
// @version 1.0
// @description REST API для управления заметками с JWT-аутентификацией.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// База данных
	db, err := postgres.New(cfg.DB.ConnectionString())
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()
	log.Println("connected to postgres")

	// Stores
	userStore := postgres.NewUserStore(db)
	noteStore := postgres.NewNoteStore(db)

	// Services
	tokenService := service.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Handlers
	authHandler := handler.NewAuthHandler(userStore, tokenService)
	noteHandler := handler.NewNoteHandler(noteStore)

	// Router
	router := handler.NewRouter(authHandler, noteHandler, tokenService)

	// HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown по SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
