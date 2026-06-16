// Package config — muhit o'zgaruvchilaridan sozlamani o'qiydi.
package config

import "os"

type Config struct {
	Port        string
	Env         string
	DatabaseURL string
	JWTSecret   string
}

// Load — env'dan sozlamani o'qiydi (oqilona standartlar bilan).
func Load() Config {
	return Config{
		Port:        getenv("PORT", "8080"),
		Env:         getenv("ENV", "development"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   getenv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
