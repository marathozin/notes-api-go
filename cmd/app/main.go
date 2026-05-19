package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/marathozin/notes-api-go/internal/handler"
	"github.com/marathozin/notes-api-go/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := &handler.Env{
		Notes: store.NewInMemoryStore(),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      handler.NewRouter(env),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
