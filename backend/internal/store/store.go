package store

import "github.com/jackc/pgx/v4/pgxpool"

// Store owns database access. Domain queries will be added alongside their features.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
