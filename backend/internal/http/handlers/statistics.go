package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgconn"
)

func (h *Handlers) ListPlayerStatistics(w http.ResponseWriter, r *http.Request) {
	statistics, err := h.store.ListPlayerStatistics(r.Context(), r.URL.Query().Get("competition"), optionalQuery(r, "player_id"))
	if err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "22P02" {
			http.Error(w, "Invalid player ID", http.StatusBadRequest)
		} else {
			log.Printf("list player statistics: %v", err)
			http.Error(w, "Could not fetch player statistics", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statistics)
}

func (h *Handlers) ListBeerBalances(w http.ResponseWriter, r *http.Request) {
	balances, err := h.store.ListBeerBalances(r.Context(), optionalQuery(r, "player_id"))
	if err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "22P02" {
			http.Error(w, "Invalid player ID", http.StatusBadRequest)
		} else {
			log.Printf("list beer balances: %v", err)
			http.Error(w, "Could not fetch beer balances", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(balances)
}
