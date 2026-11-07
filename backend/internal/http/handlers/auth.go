package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"digrone/internal/http/middleware"
	"digrone/internal/store"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const loginCookie = "digrone_login"

func (h *Handlers) setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	expires := time.Now().Add(time.Duration(maxAge) * time.Second)
	if maxAge < 0 {
		expires = time.Unix(1, 0)
	}
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/", HttpOnly: true,
		Secure:   strings.HasPrefix(h.cfg.AppURL, "https://"),
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge, Expires: expires,
	})
}

func (h *Handlers) AuthLogin(w http.ResponseWriter, r *http.Request) {
	state, nonce, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	h.setCookie(w, loginCookie, strings.Join([]string{state, nonce, verifier}, "."), 300)
	http.Redirect(w, r, h.auth.OAuth2.AuthCodeURL(state,
		oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (h *Handlers) AuthCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(loginCookie)
	h.setCookie(w, loginCookie, "", -1)
	if err != nil {
		http.Error(w, "Login expired; try again", http.StatusBadRequest)
		return
	}
	login := strings.Split(cookie.Value, ".")
	if len(login) != 3 || login[0] == "" || login[0] != r.URL.Query().Get("state") {
		http.Error(w, "Invalid login state", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		http.Error(w, "Login was not completed", http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	token, err := h.auth.OAuth2.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(login[2]))
	if err != nil {
		http.Error(w, "Keycloak login failed", http.StatusBadGateway)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "Missing identity token", http.StatusUnauthorized)
		return
	}
	identity, err := h.auth.Verifier.Verify(ctx, rawIDToken)
	if err != nil || identity.Nonce != login[1] || identity.Subject == "" {
		http.Error(w, "Invalid identity token", http.StatusUnauthorized)
		return
	}
	var claims struct {
		Username       string `json:"preferred_username"`
		Email          string `json:"email"`
		ResourceAccess map[string]struct {
			Roles []string `json:"roles"`
		} `json:"resource_access"`
	}
	if err := identity.Claims(&claims); err != nil {
		http.Error(w, "Invalid identity claims", http.StatusUnauthorized)
		return
	}
	if !slices.Contains(claims.ResourceAccess[h.cfg.OIDCClientID].Roles, h.cfg.OIDCAdminRole) {
		http.Error(w, "Admin role required", http.StatusForbidden)
		return
	}
	if claims.Username == "" {
		claims.Username = identity.Subject
	}
	// Replace any current browser session when signing in again.
	if old, err := r.Cookie(store.SessionCookie); err == nil {
		if err := h.store.DeleteSession(ctx, old.Value); err != nil {
			http.Error(w, "Could not replace session", http.StatusInternalServerError)
			return
		}
	}
	sessionToken := oauth2.GenerateVerifier()
	session := store.Session{
		Issuer: identity.Issuer, Subject: identity.Subject,
		Username: claims.Username, Email: claims.Email, IsAdmin: true,
		ExpiresAt: time.Now().Add(store.SessionLifetime),
	}
	if err := h.store.CreateSession(ctx, sessionToken, session); err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}
	h.setCookie(w, store.SessionCookie, sessionToken, int(store.SessionLifetime.Seconds()))
	http.Redirect(w, r, h.cfg.AppURL+"/admin", http.StatusFound)
}

func (h *Handlers) AuthMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(middleware.SessionFromContext(r.Context()))
}

func (h *Handlers) AuthLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(store.SessionCookie); err == nil {
		if err := h.store.DeleteSession(r.Context(), cookie.Value); err != nil {
			http.Error(w, "Could not delete session", http.StatusInternalServerError)
			return
		}
	}
	h.setCookie(w, store.SessionCookie, "", -1)
	h.setCookie(w, loginCookie, "", -1)
	w.WriteHeader(http.StatusNoContent)
}
