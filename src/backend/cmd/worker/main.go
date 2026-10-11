package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("worker started (database configured: %t)", cfg.DatabaseURL != "")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("worker shutting down")
	os.Exit(0)
}
