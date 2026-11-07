package config

import "os"

type Config struct {
	Port             string
	Version          string
	DatabaseURL      string
	AppURL           string
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCAdminRole    string
}

func FromEnv() Config {
	appURL := getEnv("APP_URL", "http://localhost:8080")
	return Config{
		Port:             getEnv("PORT", "8081"),
		Version:          getEnv("VERSION", "0.0.1"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		AppURL:           appURL,
		OIDCIssuer:       os.Getenv("OIDC_ISSUER"),
		OIDCClientID:     getEnv("OIDC_CLIENT_ID", "digrone"),
		OIDCClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:  getEnv("OIDC_REDIRECT_URL", appURL+"/auth/callback"),
		OIDCAdminRole:    getEnv("OIDC_ADMIN_ROLE", "admin"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
