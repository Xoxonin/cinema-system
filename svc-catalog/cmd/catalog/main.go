package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cinema/svc-catalog/internal/config"
	"cinema/svc-catalog/internal/httpapi"
	"cinema/svc-catalog/internal/repository/memory"
	"cinema/svc-catalog/internal/service"
)

func main() {
	cfg := config.Load()

	repo := memory.NewMovieRepository()
	svc := service.NewCatalogService(repo)
	router := httpapi.NewRouter(svc)

	server := &http.Server{
		Addr:         cfg.HTTPAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("catalog service listening on %s", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
