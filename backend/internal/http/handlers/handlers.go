package handlers

import (
	"digrone/internal/auth"
	"digrone/internal/config"
	"digrone/internal/store"
)

type Handlers struct {
	store *store.Store
	cfg   config.Config
	auth  *auth.Client
}

func NewHandlers(cfg config.Config, store *store.Store, auth *auth.Client) *Handlers {
	return &Handlers{
		store: store,
		cfg:   cfg,
		auth:  auth,
	}
}
