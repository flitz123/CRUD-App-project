package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop-scraper/internal/scraper"
	"shop-scraper/internal/store"
	"shop-scraper/internal/web"
)

func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" && os.Getenv("VERCEL") == "" {
		dataFile = "data/products.json"
	}

	catalog, err := store.New(dataFile)
	if err != nil {
		log.Fatalf("load product catalog: %v", err)
	}

	app := web.New(catalog, scraper.New(nil, ""))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("ShopScout listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("start server: %v", err)
	}
}
