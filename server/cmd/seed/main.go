// Command seed — boshlang'ich ma'lumotni Postgres'ga yozadi (idempotent).
//
// Sohalar: english (generativ), math (generativ), general (statik bank).
// english/math savollari provider tomonidan generatsiya qilinadi; general savollari
// shu yerda DB'ga yoziladi (GeneralProvider o'qiydi).
// Ishga tushirish: DATABASE_URL=... go run ./cmd/seed   (yoki `make seed`).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

var q *store.Queries
var ctx = context.Background()

func main() {
	cfg := config.Load()
	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres ulanmadi: %v", err)
	}
	defer pool.Close()
	q = store.New(pool)

	eng := ensureSubject("english", "Ingliz tili", "📘")
	ensureCategory(eng.ID, "irregular-verbs", "Noto'g'ri fe'llar")

	math := ensureSubject("math", "Matematika", "🔢")
	ensureCategory(math.ID, "arithmetic", "Arifmetika")

	gen := ensureSubject("general", "Umumiy bilim", "🌍")
	gcat := ensureCategory(gen.ID, "mixed", "Aralash")
	seedGeneral(gcat.ID)

	log.Println("seed tugadi ✓")
}

func ensureSubject(slug, name, icon string) store.Subject {
	s, err := q.GetSubjectBySlug(ctx, slug)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		s, err = q.CreateSubject(ctx, store.CreateSubjectParams{Slug: slug, Name: name, Icon: &icon})
		if err != nil {
			log.Fatalf("subject %q: %v", slug, err)
		}
		log.Printf("✓ subject: %s", slug)
	case err != nil:
		log.Fatalf("subject %q olish: %v", slug, err)
	default:
		log.Printf("• subject mavjud: %s", slug)
	}
	return s
}

func ensureCategory(subjectID uuid.UUID, slug, name string) store.Category {
	cats, err := q.ListCategoriesBySubject(ctx, subjectID)
	if err != nil {
		log.Fatalf("kategoriyalar: %v", err)
	}
	for _, c := range cats {
		if c.Slug == slug {
			log.Printf("• kategoriya mavjud: %s", slug)
			return c
		}
	}
	c, err := q.CreateCategory(ctx, store.CreateCategoryParams{SubjectID: subjectID, Slug: slug, Name: name})
	if err != nil {
		log.Fatalf("kategoriya %q: %v", slug, err)
	}
	log.Printf("✓ kategoriya: %s", slug)
	return c
}

func seedGeneral(categoryID uuid.UUID) {
	n, err := q.CountQuestionsByCategory(ctx, categoryID)
	if err != nil {
		log.Fatalf("savol soni: %v", err)
	}
	if n > 0 {
		log.Printf("• general savollar mavjud (%d)", n)
		return
	}
	questions := []store.CreateQuestionParams{
		mcqQ("Yer Quyosh atrofida necha kunda aylanadi?", "Taxminan 365 kun.",
			[][2]string{{"a", "30"}, {"b", "365"}, {"c", "24"}, {"d", "12"}}, "b"),
		mcqQ("Quyosh tizimidagi eng katta sayyora?", "Yupiter.",
			[][2]string{{"a", "Mars"}, {"b", "Yer"}, {"c", "Yupiter"}, {"d", "Venera"}}, "c"),
		mcqQ("Kattalar tanasida nechta suyak bor?", "206 ta.",
			[][2]string{{"a", "206"}, {"b", "150"}, {"c", "300"}, {"d", "100"}}, "a"),
		tfQ("Suv dengiz sathida 100°C da qaynaydi.", "To'g'ri.", true),
		tfQ("Quyosh — bu sayyora.", "Noto'g'ri, Quyosh — yulduz.", false),
	}
	for _, p := range questions {
		p.CategoryID = categoryID
		if _, err := q.CreateQuestion(ctx, p); err != nil {
			log.Fatalf("savol yozish: %v", err)
		}
	}
	log.Printf("✓ general savollar: %d", len(questions))
}

func mcqQ(prompt, expl string, opts [][2]string, correctID string) store.CreateQuestionParams {
	options := make([]map[string]string, len(opts))
	for i, o := range opts {
		options[i] = map[string]string{"id": o[0], "text": o[1]}
	}
	ob, _ := json.Marshal(options)
	cb, _ := json.Marshal(map[string]string{"optionId": correctID})
	e := expl
	return store.CreateQuestionParams{Type: "mcq", Prompt: prompt, Options: ob, Correct: cb, Explanation: &e, Difficulty: 1}
}

func tfQ(prompt, expl string, val bool) store.CreateQuestionParams {
	cb, _ := json.Marshal(map[string]bool{"value": val})
	e := expl
	return store.CreateQuestionParams{Type: "true_false", Prompt: prompt, Correct: cb, Explanation: &e, Difficulty: 1}
}
