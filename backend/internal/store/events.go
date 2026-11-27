package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
)

var ErrInvalidEvent = errors.New("invalid event: goals need a match; penalties need a match and positive minutes; fines/payments need positive num_beers and no match; other type-specific fields must be empty")

type EventInput struct {
	MatchID           *string    `json:"match_id"`
	PlayerID          string     `json:"player_id"`
	EventType         string     `json:"event_type"`
	AssistingPlayerID *string    `json:"assisting_player_id"`
	PenaltyMinutes    *int       `json:"penalty_minutes"`
	NumBeers          *int       `json:"num_beers"`
	Comment           *string    `json:"comment"`
	OccurredAt        *time.Time `json:"occurred_at"`
}

type Event struct {
	ID                string    `json:"id"`
	MatchID           *string   `json:"match_id"`
	PlayerID          string    `json:"player_id"`
	EventType         string    `json:"event_type"`
	AssistingPlayerID *string   `json:"assisting_player_id"`
	PenaltyMinutes    *int      `json:"penalty_minutes"`
	NumBeers          *int      `json:"num_beers"`
	Comment           *string   `json:"comment"`
	OccurredAt        time.Time `json:"occurred_at"`
}

func (input EventInput) validate() error {
	hasMatch := input.MatchID != nil && *input.MatchID != ""
	valid := input.PlayerID != ""
	switch input.EventType {
	case "goal":
		valid = valid && hasMatch && input.PenaltyMinutes == nil && input.NumBeers == nil
	case "penalty":
		valid = valid && hasMatch && input.PenaltyMinutes != nil && *input.PenaltyMinutes > 0 && input.NumBeers == nil && input.AssistingPlayerID == nil
	case "fine", "fine_payment":
		valid = valid && input.MatchID == nil && input.NumBeers != nil && *input.NumBeers > 0 && input.PenaltyMinutes == nil && input.AssistingPlayerID == nil
	default:
		valid = false
	}
	if !valid {
		return ErrInvalidEvent
	}
	return nil
}

type EventFilter struct {
	MatchID   *string
	PlayerID  *string
	EventType string
}

// Both the connection pool and a match-save transaction implement this interface.
type eventDB interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

const eventColumns = "id, match_id, player_id, event_type, assisting_player_id, penalty_minutes, num_beers, comment, occurred_at"

func scanEvent(row pgx.Row) (Event, error) {
	var event Event
	err := row.Scan(&event.ID, &event.MatchID, &event.PlayerID, &event.EventType,
		&event.AssistingPlayerID, &event.PenaltyMinutes, &event.NumBeers, &event.Comment, &event.OccurredAt)
	return event, err
}

func listEvents(ctx context.Context, db eventDB, filter EventFilter) ([]Event, error) {
	rows, err := db.Query(ctx, "SELECT "+eventColumns+` FROM public.events
 WHERE ($1::uuid IS NULL OR match_id = $1)
 AND ($2::uuid IS NULL OR player_id = $2)
 AND ($3 = '' OR event_type = $3)
 ORDER BY occurred_at, id`, filter.MatchID, filter.PlayerID, filter.EventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) ListEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	return listEvents(ctx, s.pool, filter)
}

func (s *Store) GetEvent(ctx context.Context, id string) (Event, error) {
	return scanEvent(s.pool.QueryRow(ctx, "SELECT "+eventColumns+" FROM public.events WHERE id = $1", id))
}

func insertEvent(ctx context.Context, db eventDB, input EventInput) (Event, error) {
	if err := input.validate(); err != nil {
		return Event{}, err
	}
	return scanEvent(db.QueryRow(ctx, `INSERT INTO public.events
 (match_id, player_id, event_type, assisting_player_id, penalty_minutes, num_beers, comment, occurred_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, now())) RETURNING `+eventColumns,
		input.MatchID, input.PlayerID, input.EventType, input.AssistingPlayerID,
		input.PenaltyMinutes, input.NumBeers, input.Comment, input.OccurredAt))
}

func (s *Store) CreateEvent(ctx context.Context, input EventInput) (Event, error) {
	return insertEvent(ctx, s.pool, input)
}

func (s *Store) UpdateEvent(ctx context.Context, id string, input EventInput) (Event, error) {
	if err := input.validate(); err != nil {
		return Event{}, err
	}
	return scanEvent(s.pool.QueryRow(ctx, `UPDATE public.events SET
 match_id=$2, player_id=$3, event_type=$4, assisting_player_id=$5,
 penalty_minutes=$6, num_beers=$7, comment=$8, occurred_at=COALESCE($9, occurred_at)
 WHERE id=$1 RETURNING `+eventColumns,
		id, input.MatchID, input.PlayerID, input.EventType, input.AssistingPlayerID,
		input.PenaltyMinutes, input.NumBeers, input.Comment, input.OccurredAt))
}

func (s *Store) DeleteEvent(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM public.events WHERE id=$1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
