package routes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/toluhikay/fx-exchange/internal/fx"
	"github.com/toluhikay/fx-exchange/internal/handlers"
	fxMiddleware "github.com/toluhikay/fx-exchange/internal/middleware"
	"github.com/toluhikay/fx-exchange/internal/repository"
	"github.com/toluhikay/fx-exchange/internal/services"
	"github.com/toluhikay/fx-exchange/pkg/jwt"
)

type RouteConfig struct {
	db               *sql.DB
	ctx              context.Context
	fxProvider       fx.FXProvider
	customMiddleware fxMiddleware.AuthMiddleware
	auth             jwt.Auth
}

func NewRouteConfig(db *sql.DB, ctx context.Context, fxProvider fx.FXProvider, auth jwt.Auth, customMiddleWare fxMiddleware.AuthMiddleware) *RouteConfig {
	return &RouteConfig{
		db:               db,
		ctx:              ctx,
		fxProvider:       fxProvider,
		auth:             auth,
		customMiddleware: customMiddleWare,
	}
}

func (r *RouteConfig) SetUpRoutes() http.Handler {

	repo := repository.NewRepository(r.db)
	userRepo := repository.NewUserRepo(r.db)
	if os.Getenv("USE_MOCK_FX") == "true" {
		r.fxProvider = fx.NewMockFXProvider(true)
	} else {
		r.fxProvider = fx.NewFXClient()
	}

	fmt.Println("fx provider", r.fxProvider)

	svc := services.NewService(repo, r.fxProvider)
	userSvc := services.NewUserService(*userRepo)

	handler := handlers.NewHandler(svc, userSvc)
	wsHandler := handlers.NewWebSocketHandler(r.fxProvider)
	userHandlers := handlers.NewUserHandler(*userSvc, r.auth)

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://fx-exchange-front.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	mux.Use(corsMiddleware.Handler)

	// register handlers here
	mux.Post("/api/user/register", userHandlers.CreateUser)
	mux.Post("/api/user/login", userHandlers.UserLogin)

	mux.Route("/api/user", func(mux chi.Router) {
		mux.Get("/api/user/", userHandlers.GetUserById)
	})

	mux.Route("/api/wallets", func(mux chi.Router) {
		mux.Use(r.customMiddleware.AuthRequired)
		mux.Post("/", handler.CreateWallet)
		mux.Get("/", handler.GetWallet)
		mux.Post("/deposit", handler.Deposit)
		mux.Get("/balances", handler.GetBalances)
		mux.Post("/swap", handler.Swap)
		mux.Post("/transfer", handler.Transfer)
		mux.Get("/history", handler.GetTransactionHistory)
	})

	mux.Get("/ws/fx-rates", wsHandler.HandleFXRates)

	return mux
}
