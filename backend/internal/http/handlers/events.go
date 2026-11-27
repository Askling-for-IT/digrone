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

func optionalQuery(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}

func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.store.ListEvents(r.Context(), store.EventFilter{
		MatchID:   optionalQuery(r, "match_id"),
		PlayerID:  optionalQuery(r, "player_id"),
		EventType: r.URL.Query().Get("event_type"),
	})
	if err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("list events: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(events)
}

func (h *Handlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	event, err := h.store.GetEvent(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Not found", http.StatusNotFound)
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("get event: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(event)
}

func (h *Handlers) SaveEvent(w http.ResponseWriter, r *http.Request) {
	var input store.EventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Could not read event", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	var event store.Event
	var err error
	status := http.StatusOK
	if id == "" {
		event, err = h.store.CreateEvent(r.Context(), input)
		status = http.StatusCreated
	} else {
		event, err = h.store.UpdateEvent(r.Context(), id, input)
	}
	if err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Not found", http.StatusNotFound)
		case errors.Is(err, store.ErrInvalidEvent):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.As(err, &databaseError) && databaseError.Code == "23503":
			http.Error(w, "Referenced player or match does not exist", http.StatusConflict)
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("save event: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	if status == http.StatusCreated {
		w.Header().Set("Location", "/api/events/"+event.ID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(event)
}

func (h *Handlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteEvent(r.Context(), chi.URLParam(r, "id")); err != nil {
		var databaseError *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "Not found", http.StatusNotFound)
		case errors.As(err, &databaseError) && (databaseError.Code == "23514" || databaseError.Code == "22P02"):
			http.Error(w, "Invalid match or event fields", http.StatusBadRequest)
		default:
			log.Printf("delete event: %v", err)
			http.Error(w, "Could not save or fetch data", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
