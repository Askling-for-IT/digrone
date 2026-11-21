package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"digrone/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func (h *Handlers) ListMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := h.store.ListMatches(r.Context(), r.URL.Query().Get("competition"))
	if err != nil {
		log.Printf("list matches: %v", err)
		http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(matches)
}

func (h *Handlers) GetMatch(w http.ResponseWriter, r *http.Request) {
	match, err := h.store.GetMatch(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Not found", http.StatusNotFound)
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("get match: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(match)
}

func (h *Handlers) SaveMatch(w http.ResponseWriter, r *http.Request) {
	var input store.MatchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Could not read match", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	match, err := h.store.SaveMatch(r.Context(), id, input)
	if err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Not found", http.StatusNotFound)
		case errors.Is(err, store.ErrInvalidEvent), errors.Is(err, store.ErrInvalidMatch):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.As(err, &databaseError) && databaseError.Code == "23503":
			http.Error(w, "Referenced player or match does not exist", http.StatusConflict)
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("save match: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
		w.Header().Set("Location", "/api/matches/"+match.ID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(match)
}
