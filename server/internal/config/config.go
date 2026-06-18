// Package config — muhit o'zgaruvchilaridan sozlamani o'qiydi.
package config

import "os"

type Config struct {
	Port             string
	Env              string
	DatabaseURL      string
	JWTSecret        string
	TelegramBotToken string // "" bo'lsa Telegram auth + bot o'chiq
	MiniAppURL       string // bot /start tugmasi shu URL'ni ochadi
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
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
