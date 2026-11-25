package store

import "context"

type PlayerStatistics struct {
	PlayerID       string `json:"player_id"`
	Name           string `json:"name"`
	Goals          int64  `json:"goals"`
	Assists        int64  `json:"assists"`
	Points         int64  `json:"points"`
	PenaltyMinutes int64  `json:"penalty_minutes"`
}

type BeerBalance struct {
	PlayerID         string `json:"player_id"`
	Name             string `json:"name"`
	FinedBeers       int64  `json:"fined_beers"`
	PaidBeers        int64  `json:"paid_beers"`
	OutstandingBeers int64  `json:"outstanding_beers"`
}

func (s *Store) ListPlayerStatistics(ctx context.Context, competition string, playerID *string) ([]PlayerStatistics, error) {
	// Each aggregate has one row per player, preventing scorer/assister joins
	// from multiplying goals, assists, or penalty minutes.
	rows, err := s.pool.Query(ctx, `
WITH sporting_events AS (
    SELECT e.player_id, e.event_type, e.assisting_player_id, e.penalty_minutes
    FROM public.events e
    JOIN public.matches m ON m.id = e.match_id
    WHERE m.status = 'completed'
      AND ($1 = '' OR m.competition = $1)
      AND e.event_type IN ('goal', 'penalty')
), player_totals AS (
    SELECT player_id,
           COUNT(*) FILTER (WHERE event_type = 'goal') AS goals,
           COALESCE(SUM(penalty_minutes) FILTER (WHERE event_type = 'penalty'), 0) AS penalty_minutes
    FROM sporting_events
    GROUP BY player_id
), assist_totals AS (
    SELECT assisting_player_id AS player_id, COUNT(*) AS assists
    FROM sporting_events
    WHERE event_type = 'goal' AND assisting_player_id IS NOT NULL
    GROUP BY assisting_player_id
)
SELECT p.id, p.name,
       COALESCE(t.goals, 0), COALESCE(a.assists, 0),
       COALESCE(t.goals, 0) + COALESCE(a.assists, 0) AS points,
       COALESCE(t.penalty_minutes, 0)
FROM public.players p
LEFT JOIN player_totals t ON t.player_id = p.id
LEFT JOIN assist_totals a ON a.player_id = p.id
WHERE ($2::uuid IS NULL OR p.id = $2)
ORDER BY points DESC, COALESCE(t.goals, 0) DESC, p.name, p.id`, competition, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	statistics := make([]PlayerStatistics, 0)
	for rows.Next() {
		var player PlayerStatistics
		if err := rows.Scan(&player.PlayerID, &player.Name, &player.Goals, &player.Assists, &player.Points, &player.PenaltyMinutes); err != nil {
			return nil, err
		}
		statistics = append(statistics, player)
	}
	return statistics, rows.Err()
}

func (s *Store) ListBeerBalances(ctx context.Context, playerID *string) ([]BeerBalance, error) {
	rows, err := s.pool.Query(ctx, `
WITH beer_totals AS (
    SELECT player_id,
           COALESCE(SUM(num_beers) FILTER (WHERE event_type = 'fine'), 0) AS fined_beers,
           COALESCE(SUM(num_beers) FILTER (WHERE event_type = 'fine_payment'), 0) AS paid_beers
    FROM public.events
    WHERE event_type IN ('fine', 'fine_payment')
    GROUP BY player_id
)
SELECT p.id, p.name, COALESCE(t.fined_beers, 0), COALESCE(t.paid_beers, 0),
       COALESCE(t.fined_beers, 0) - COALESCE(t.paid_beers, 0) AS outstanding_beers
FROM public.players p
LEFT JOIN beer_totals t ON t.player_id = p.id
WHERE ($1::uuid IS NULL OR p.id = $1)
ORDER BY outstanding_beers DESC, p.name, p.id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	balances := make([]BeerBalance, 0)
	for rows.Next() {
		var player BeerBalance
		if err := rows.Scan(&player.PlayerID, &player.Name, &player.FinedBeers, &player.PaidBeers, &player.OutstandingBeers); err != nil {
			return nil, err
		}
		balances = append(balances, player)
	}
	return balances, rows.Err()
}
