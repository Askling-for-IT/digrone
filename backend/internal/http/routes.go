package http

import (
	"encoding/json"
	"log"
	"net/http"

	"digrone/internal/auth"
	"digrone/internal/config"
	"digrone/internal/http/handlers"
	appmw "digrone/internal/http/middleware"
	"digrone/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool, auth *auth.Client) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	// Avoid logging callback query strings containing authorization codes.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Recoverer)
	r.Use(http.NewCrossOriginProtection().Handler)

	s := store.NewStore(pool)
	h := handlers.NewHandlers(cfg, s, auth)
	r.Route("/api/players", func(r chi.Router) {
		r.Get("/", h.ListPlayers)
		r.Get("/{id}", h.GetPlayer)
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAdmin(s))
			r.Post("/", h.CreatePlayer)
			r.Put("/{id}", h.UpdatePlayer)
			r.Delete("/{id}", h.DeletePlayer)
		})
	})
	r.Route("/auth", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-store")
				w.Header().Set("Referrer-Policy", "no-referrer")
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/login", h.AuthLogin)
		r.Get("/callback", h.AuthCallback)
		r.With(appmw.RequireAdmin(s)).Get("/me", h.AuthMe)
		r.Post("/logout", h.AuthLogout)
	})
	r.Get("/health", handlers.Health)
	r.Get("/db-health", h.DbHealth)
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Version string `json:"version"`
		}{Version: cfg.Version})
	})
	return r
}
