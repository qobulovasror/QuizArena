// Command setadmin — foydalanuvchini admin qiladi (role=admin).
// Ishlatish: DATABASE_URL=... go run ./cmd/setadmin -email=ali@example.com
package main

import (
	"context"
	"flag"
	"log"

	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

func main() {
	email := flag.String("email", "", "admin qilinadigan foydalanuvchi email")
	flag.Parse()
	if *email == "" {
		log.Fatal("-email kerak")
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	if err := store.New(pool).SetRoleByEmail(ctx, store.SetRoleByEmailParams{Email: email, Role: "admin"}); err != nil {
		log.Fatalf("rol yangilash: %v", err)
	}
	log.Printf("✓ %s endi admin", *email)
}
