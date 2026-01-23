package api

import "github.com/Nevnet99/trade-engine/internal/store"

type Server struct {
	store *store.Storage
}

func NewServer(store *store.Storage) *Server {
	return &Server{
		store: store,
	}
}
