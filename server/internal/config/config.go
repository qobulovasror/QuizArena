// Package config — muhit o'zgaruvchilaridan sozlamani o'qiydi.
package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	Env              string
	DatabaseURL      string
	JWTSecret        string
	TelegramBotToken string   // "" bo'lsa Telegram auth + bot o'chiq
	MiniAppURL       string   // bot /start tugmasi shu URL'ni ochadi
	CORSOrigins      []string // bo'sh → hammaga ruxsat (dev); prod'da ro'yxat
}

// Load — env'dan sozlamani o'qiydi (oqilona standartlar bilan).
func Load() Config {
	return Config{
		Port:             getenv("PORT", "8080"),
		Env:              getenv("ENV", "development"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        getenv("JWT_SECRET", "dev-secret-change-me"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		MiniAppURL:       os.Getenv("MINIAPP_URL"),
		CORSOrigins:      splitCSV(os.Getenv("CORS_ORIGINS")),
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
