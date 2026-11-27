package middleware

import (
	"context"
	"errors"
	"net/http"

	"digrone/internal/store"
	"github.com/jackc/pgx/v4"
)

type sessionKey struct{}

func SessionFromContext(ctx context.Context) store.Session {
	return ctx.Value(sessionKey{}).(store.Session)
}

// RequireAdmin protects the current-user endpoint and administrative write routes.
func RequireAdmin(s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			cookie, err := r.Cookie(store.SessionCookie)
			if err != nil {
				http.Error(w, "Login required", http.StatusUnauthorized)
				return
			}
			session, err := s.GetSession(r.Context(), cookie.Value)
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Login required", http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, "Could not load session", http.StatusInternalServerError)
				return
			}
			if !session.IsAdmin {
				http.Error(w, "Admin role required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, session)))
		})
	}
}
