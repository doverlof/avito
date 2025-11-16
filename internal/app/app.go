package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/doverlof/avito_help/api"
	pullRequestRepoPkg "github.com/doverlof/avito_help/internal/client/repo/pull-request"
	statsRepoPkg "github.com/doverlof/avito_help/internal/client/repo/pull-request"
	teamRepoPkg "github.com/doverlof/avito_help/internal/client/repo/team"
	userRepoPkg "github.com/doverlof/avito_help/internal/client/repo/user"

	"github.com/doverlof/avito_help/internal/config"
	"github.com/doverlof/avito_help/internal/handler"
	pullRequestUsecasePkg "github.com/doverlof/avito_help/internal/usecase/pull-request"
	statsUseCasePkg "github.com/doverlof/avito_help/internal/usecase/stats"
	teamUseCasePkg "github.com/doverlof/avito_help/internal/usecase/team"
	userUseCasePkg "github.com/doverlof/avito_help/internal/usecase/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func MustConfigureApp(r *chi.Mux, cfg *config.Config) func(ctx context.Context) {
	fmt.Println("Initializing app")

	//Clients
	fmt.Println("Init postgres")
	sqlClient := initPostgresClient(&cfg.PostgresConfig)

	//Repos
	teamRepo := teamRepoPkg.New(sqlClient)
	pullRequestRepo := pullRequestRepoPkg.New(sqlClient)
	userRepo := userRepoPkg.New(sqlClient)
	statsRepo := statsRepoPkg.New(sqlClient)

	//UseCases

	teamUseCase := teamUseCasePkg.New(teamRepo)
	pullRequestUseCase := pullRequestUsecasePkg.New(pullRequestRepo, userRepo)

	userUseCase := userUseCasePkg.New(userRepo)
	statsUseCase := statsUseCasePkg.New(statsRepo)
	//Handlers

	fmt.Println("Create server")
	server := handler.New(teamUseCase, userUseCase, statsUseCase, pullRequestUseCase)

	//Middleware

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AllowOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//HTTP handler
	httpHandler := api.HandlerWithOptions(server, api.ChiServerOptions{
		BaseRouter: r,
	})

	//Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.RestConfig.Port),
		Handler: httpHandler,
	}

	go func() {
		log.Printf("Starting HTTP server on port %d\n", cfg.RestConfig.Port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start HTTP server", err)
		}
	}()

	return func(ctx context.Context) {
		if err := sqlClient.Close(); err != nil {
			fmt.Println(err)
		}
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Println(err)
		}
	}
}
