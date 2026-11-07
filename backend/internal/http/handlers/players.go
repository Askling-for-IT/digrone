package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type playerRequest struct {
	Name string `json:"name"`
}

func playerError(w http.ResponseWriter, err error) {
	var databaseError *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		http.Error(w, "Player not found", http.StatusNotFound)
	case errors.As(err, &databaseError) && databaseError.Code == "23503":
		http.Error(w, "Player is referenced by events", http.StatusConflict)
	default:
		log.Printf("player database error: %v", err)
		http.Error(w, "Could not access players", http.StatusInternalServerError)
	}
}

func (h *Handlers) ListPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := h.store.ListPlayers(r.Context())
	if err != nil {
		playerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(players)
}

func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	player, err := h.store.GetPlayer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		playerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(player)
}

func (h *Handlers) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var input playerRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Could not read player", http.StatusBadRequest)
		return
	}
	player, err := h.store.CreatePlayer(r.Context(), input.Name)
	if err != nil {
		playerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/players/"+player.ID)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(player)
}

func (h *Handlers) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	var input playerRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Could not read player", http.StatusBadRequest)
		return
	}
	player, err := h.store.UpdatePlayer(r.Context(), chi.URLParam(r, "id"), input.Name)
	if err != nil {
		playerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(player)
}

func (h *Handlers) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeletePlayer(r.Context(), chi.URLParam(r, "id")); err != nil {
		playerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
