package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const SessionCookie = "digrone_session"
const SessionLifetime = 8 * time.Hour

type Session struct {
	Issuer    string    `json:"-"`
	Subject   string    `json:"subject"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	ExpiresAt time.Time `json:"expires_at"`
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Store) CreateSession(ctx context.Context, token string, session Session) error {
	// Expired sessions are rejected on reads and removed on subsequent logins.
	if _, err := s.pool.Exec(ctx, "DELETE FROM public.sessions WHERE expires_at <= now()"); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
  INSERT INTO public.sessions (token_hash, issuer, subject, username, email, is_admin, expires_at)
  VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tokenHash(token), session.Issuer, session.Subject, session.Username,
		session.Email, session.IsAdmin, session.ExpiresAt,
	)
	return err
}

func (s *Store) GetSession(ctx context.Context, token string) (Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
  SELECT issuer, subject, username, email, is_admin, expires_at
  FROM public.sessions WHERE token_hash = $1 AND expires_at > now()`, tokenHash(token),
	).Scan(&session.Issuer, &session.Subject, &session.Username, &session.Email, &session.IsAdmin, &session.ExpiresAt)
	return session, err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM public.sessions WHERE token_hash = $1", tokenHash(token))
	return err
}
