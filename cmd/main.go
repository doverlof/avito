package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doverlof/avito_help/internal/app"
	"github.com/doverlof/avito_help/internal/config"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoadConfig()

	r := chi.NewRouter()

	closeFn := app.MustConfigureApp(r, cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	closeFn(ctx)

	log.Println("Server stopped")
}
