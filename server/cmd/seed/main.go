// Command seed — boshlang'ich ma'lumotni Postgres'ga yozadi (idempotent).
//
// Hozircha: english soha + irregular-verbs kategoriya rejatlari (savollar
// EnglishVerbProvider tomonidan generatsiya qilinadi, bankda saqlanmaydi).
// Ishga tushirish: DATABASE_URL=... go run ./cmd/seed   (yoki `make seed`).
package main

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres ulanmadi: %v", err)
	}
	defer pool.Close()
	q := store.New(pool)

	// english soha (idempotent)
	subj, err := q.GetSubjectBySlug(ctx, "english")
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		icon := "📘"
		subj, err = q.CreateSubject(ctx, store.CreateSubjectParams{Slug: "english", Name: "Ingliz tili", Icon: &icon})
		if err != nil {
			log.Fatalf("subject yaratish: %v", err)
		}
		log.Printf("✓ subject yaratildi: english (%s)", subj.ID)
	case err != nil:
		log.Fatalf("subject olish: %v", err)
	default:
		log.Printf("• subject mavjud: english (%s)", subj.ID)
	}

	// irregular-verbs kategoriya (idempotent)
	cats, err := q.ListCategoriesBySubject(ctx, subj.ID)
	if err != nil {
		log.Fatalf("kategoriyalarni olish: %v", err)
	}
	exists := false
	for _, c := range cats {
		if c.Slug == "irregular-verbs" {
			exists = true
		}
	}
	if exists {
		log.Printf("• kategoriya mavjud: irregular-verbs")
	} else {
		c, err := q.CreateCategory(ctx, store.CreateCategoryParams{
			SubjectID: subj.ID, Slug: "irregular-verbs", Name: "Noto'g'ri fe'llar",
		})
		if err != nil {
			log.Fatalf("kategoriya yaratish: %v", err)
		}
		log.Printf("✓ kategoriya yaratildi: irregular-verbs (%s)", c.ID)
	}

	log.Println("seed tugadi ✓")
}
