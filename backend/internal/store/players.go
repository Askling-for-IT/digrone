package store

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Store) ListPlayers(ctx context.Context) ([]Player, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name FROM public.players ORDER BY name, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name); err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, rows.Err()
}

func (s *Store) GetPlayer(ctx context.Context, id string) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, "SELECT id, name FROM public.players WHERE id = $1", id).
		Scan(&player.ID, &player.Name)
	return player, err
}

func (s *Store) CreatePlayer(ctx context.Context, name string) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, "INSERT INTO public.players (name) VALUES ($1) RETURNING id, name", name).
		Scan(&player.ID, &player.Name)
	return player, err
}

func (s *Store) UpdatePlayer(ctx context.Context, id, name string) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, "UPDATE public.players SET name = $2 WHERE id = $1 RETURNING id, name", id, name).
		Scan(&player.ID, &player.Name)
	return player, err
}

func (s *Store) DeletePlayer(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM public.players WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
