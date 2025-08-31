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

	"github.com/joho/godotenv"
	"github.com/toluhikay/fx-exchange/internal/config"
	"github.com/toluhikay/fx-exchange/internal/db"
	"github.com/toluhikay/fx-exchange/internal/fx"
	"github.com/toluhikay/fx-exchange/internal/middleware"
	"github.com/toluhikay/fx-exchange/internal/routes"
	"github.com/toluhikay/fx-exchange/pkg/jwt"
)

func main() {
	// godotenv.Load()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading env")
	}

	cfg := config.LoadConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName,
	)
	fxProvider := fx.NewMockFXProvider(true)
	auth := jwt.NewAuth(cfg.Auth)
	customMiddleware := middleware.NewMiddleware(&cfg.Auth)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbInstance, err := db.ConnectDb(dsn)
	if err != nil {
		log.Fatal("error connecting to db: ", err)
	}

	fmt.Println("db connected successfuly")

	go fxProvider.StartRateUpdates(ctx, dbInstance, 5*time.Hour)

	routesInstance := routes.NewRouteConfig(dbInstance, ctx, fxProvider, *auth, customMiddleware)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: routesInstance.SetUpRoutes(),
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	fmt.Printf("Starting server on port %s\n", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("failed to start server: %w", err))
	}
}
