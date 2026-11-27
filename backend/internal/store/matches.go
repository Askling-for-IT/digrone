package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
)

var ErrInvalidMatch = errors.New("completed matches require goals_for and goals_against")

type MatchFields struct {
	StartTime    time.Time `json:"start_time"`
	Opponent     string    `json:"opponent"`
	HomeGame     bool      `json:"home_game"`
	Competition  string    `json:"competition"`
	Status       string    `json:"status"`
	GoalsFor     *int      `json:"goals_for"`
	GoalsAgainst *int      `json:"goals_against"`
}

type Match struct {
	ID string `json:"id"`
	MatchFields
	StatisticsInHistoricalTotals bool `json:"statistics_in_historical_totals"`
}

type MatchInput struct {
	MatchFields
	// Omitted/null preserves events; an explicit array replaces them (including []).
	Events *[]EventInput `json:"events"`
}

type MatchDetail struct {
	Match
	Events []Event `json:"events"`
}

const matchColumns = "id, start_time, opponent, home_game, competition, status, goals_for, goals_against, historical_source_key IS NOT NULL AS statistics_in_historical_totals"

func scanMatch(row pgx.Row) (Match, error) {
	var match Match
	err := row.Scan(&match.ID, &match.StartTime, &match.Opponent, &match.HomeGame,
		&match.Competition, &match.Status, &match.GoalsFor, &match.GoalsAgainst, &match.StatisticsInHistoricalTotals)
	return match, err
}

func (s *Store) ListMatches(ctx context.Context, competition string) ([]Match, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+matchColumns+` FROM public.matches
 WHERE ($1 = '' OR competition = $1) ORDER BY start_time, id`, competition)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make([]Match, 0)
	for rows.Next() {
		match, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, rows.Err()
}

func (s *Store) GetMatch(ctx context.Context, id string) (MatchDetail, error) {
	// Read the result and its events from the same snapshot during combined saves.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return MatchDetail{}, err
	}
	defer tx.Rollback(ctx)
	match, err := scanMatch(tx.QueryRow(ctx, "SELECT "+matchColumns+" FROM public.matches WHERE id=$1", id))
	if err != nil {
		return MatchDetail{}, err
	}
	events, err := listEvents(ctx, tx, EventFilter{MatchID: &id})
	if err != nil {
		return MatchDetail{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MatchDetail{}, err
	}
	return MatchDetail{Match: match, Events: events}, nil
}

// SaveMatch creates (empty id) or updates a match and its supplied events atomically.
func (s *Store) SaveMatch(ctx context.Context, id string, input MatchInput) (MatchDetail, error) {
	if input.Status == "" {
		input.Status = "scheduled"
	}
	if input.Status == "completed" && (input.GoalsFor == nil || input.GoalsAgainst == nil) {
		return MatchDetail{}, ErrInvalidMatch
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MatchDetail{}, err
	}
	defer tx.Rollback(ctx)
	var match Match
	if id == "" {
		match, err = scanMatch(tx.QueryRow(ctx, `INSERT INTO public.matches
   (start_time, opponent, home_game, competition, status, goals_for, goals_against)
   VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING `+matchColumns,
			input.StartTime, input.Opponent, input.HomeGame, input.Competition, input.Status, input.GoalsFor, input.GoalsAgainst))
	} else {
		match, err = scanMatch(tx.QueryRow(ctx, `UPDATE public.matches SET
   start_time=$2, opponent=$3, home_game=$4, competition=$5, status=$6, goals_for=$7, goals_against=$8
   WHERE id=$1 RETURNING `+matchColumns,
			id, input.StartTime, input.Opponent, input.HomeGame, input.Competition, input.Status, input.GoalsFor, input.GoalsAgainst))
	}
	if err != nil {
		return MatchDetail{}, err
	}
	if input.Events != nil {
		if _, err := tx.Exec(ctx, "DELETE FROM public.events WHERE match_id=$1", match.ID); err != nil {
			return MatchDetail{}, err
		}
		for _, event := range *input.Events {
			// All nested events belong to this match. This also rejects fines/payments.
			event.MatchID = &match.ID
			if _, err := insertEvent(ctx, tx, event); err != nil {
				return MatchDetail{}, err
			}
		}
	}
	events, err := listEvents(ctx, tx, EventFilter{MatchID: &match.ID})
	if err != nil {
		return MatchDetail{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MatchDetail{}, err
	}
	return MatchDetail{Match: match, Events: events}, nil
}
