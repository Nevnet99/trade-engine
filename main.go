package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Nevnet99/trade-engine/internal/api"
	"github.com/Nevnet99/trade-engine/internal/engine"
	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		log.Fatal("Unable to connect to database", err)
	}

	defer pool.Close()

	storage := store.NewStorageFromPool(pool)
	server := api.NewServer(storage)

	matchingEngine := engine.New(storage)

	slog.Info("Starting Matching Engine...")
	go matchingEngine.ProcessMatches(context.Background())

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Public

	r.Get("/pairs", server.HandleGetPairs)
	r.Get("/orderbook", server.HandleGetOrderBook)
	r.Get("/trades", server.HandleGetRecentTrades)
	r.Get("/kline", server.HandleGetKlines)

	// Authentication

	r.Post("/register", server.HandleCreateUser)
	r.Post("/login", server.HandleLoginUser)

	// Protected

	r.Group(func(r chi.Router) {
		r.Use(server.AuthMiddleware)

		r.Post("/trade", server.CreateOrder)
	})

	slog.Info("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
