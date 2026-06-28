// Command seed — boshlang'ich ma'lumotni Postgres'ga yozadi (idempotent).
//
// Sohalar: english (generativ), math (generativ), general + programming (statik bank).
// english/math savollari provider tomonidan generatsiya qilinadi; general savollari
// shu yerda DB'ga yoziladi (GeneralProvider o'qiydi).
// Ishga tushirish: DATABASE_URL=... go run ./cmd/seed   (yoki `make seed`).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/game/providers"
	"github.com/azizbek12234/quizarena/server/internal/state"
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

	// english — generativ provider'dan DB bankka (SRS/Baholash uchun)
	eng := ensureSubject("english", "Ingliz tili", "📘")
	engCat := ensureCategory(eng.ID, "irregular-verbs", "Noto'g'ri fe'llar")
	if ev, err := providers.NewEnglishVerb(); err == nil {
		qs, _ := ev.Questions(60)
		seedProvider(engCat.ID, qs)
	}

	// math — generativ provider'dan DB bankka
	math := ensureSubject("math", "Matematika", "🔢")
	mathCat := ensureCategory(math.ID, "arithmetic", "Arifmetika")
	mqs, _ := providers.NewMath().Questions(50)
	seedProvider(mathCat.ID, mqs)

	gen := ensureSubject("general", "Umumiy bilim", "🌍")
	gcat := ensureCategory(gen.ID, "mixed", "Aralash")
	seedGeneral(gcat.ID)

	// programming — statik bank (terminologiya, kod natijasi). General provider o'qiydi.
	prog := ensureSubject("programming", "Dasturlash", "💻")
	pcat := ensureCategory(prog.ID, "fundamentals", "Asoslar")
	seedProgramming(pcat.ID)

	log.Println("seed tugadi ✓")
}

// seedProvider — generativ provider savollarini DB'ga yozadi (idempotent, takrorsiz prompt).
// Variantlar tartibi aralashtiriladi (to'g'ri javob doim birinchi turmasligi uchun).
func seedProvider(categoryID uuid.UUID, qs []state.Question) {
	n, err := q.CountQuestionsByCategory(ctx, categoryID)
	if err != nil {
		log.Fatalf("savol soni: %v", err)
	}
	if n > 0 {
		log.Printf("• savollar mavjud (%d)", n)
		return
	}
	seen := map[string]bool{}
	count := 0
	for _, item := range qs {
		if seen[item.Prompt] {
			continue
		}
		seen[item.Prompt] = true

		var optsJSON []byte
		if len(item.Options) > 0 {
			opts := append([]state.Option(nil), item.Options...)
			rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
			optsJSON, _ = json.Marshal(opts)
		}
		e := item.Explanation
		if _, err := q.CreateQuestion(ctx, store.CreateQuestionParams{
			CategoryID: categoryID, Type: item.Type, Prompt: item.Prompt,
			Options: optsJSON, Correct: item.Correct, Explanation: &e, Difficulty: 1,
		}); err != nil {
			log.Fatalf("savol yozish: %v", err)
		}
		count++
	}
	log.Printf("✓ %d savol seed qilindi", count)
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

func seedProgramming(categoryID uuid.UUID) {
	n, err := q.CountQuestionsByCategory(ctx, categoryID)
	if err != nil {
		log.Fatalf("savol soni: %v", err)
	}
	if n > 0 {
		log.Printf("• programming savollar mavjud (%d)", n)
		return
	}
	questions := []store.CreateQuestionParams{
		mcqQ("HTTP qaysi portda standart ishlaydi?", "HTTP — 80, HTTPS — 443.",
			[][2]string{{"a", "21"}, {"b", "80"}, {"c", "443"}, {"d", "8080"}}, "b"),
		mcqQ("Qaysi tuzilma «FIFO» (birinchi kirgan birinchi chiqadi)?", "Queue — FIFO; Stack — LIFO.",
			[][2]string{{"a", "Stack"}, {"b", "Queue"}, {"c", "Tree"}, {"d", "Graph"}}, "b"),
		mcqQ("Binary qidiruvning vaqt murakkabligi?", "Saralangan massivda O(log n).",
			[][2]string{{"a", "O(n)"}, {"b", "O(n²)"}, {"c", "O(log n)"}, {"d", "O(1)"}}, "c"),
		mcqQ("Git'da o'zgarishlarni tasdiqlash buyrug'i?", "git commit.",
			[][2]string{{"a", "git push"}, {"b", "git commit"}, {"c", "git stage"}, {"d", "git save"}}, "b"),
		mcqQ("`SELECT` SQL buyrug'i nima qiladi?", "Ma'lumotni o'qiydi (o'zgartirmaydi).",
			[][2]string{{"a", "O'chiradi"}, {"b", "Yangilaydi"}, {"c", "O'qiydi"}, {"d", "Qo'shadi"}}, "c"),
		tfQ("Stack «LIFO» (oxirgi kirgan birinchi chiqadi) tamoyilida ishlaydi.", "To'g'ri.", true),
		tfQ("HTML — bu dasturlash tili.", "Noto'g'ri, HTML — belgilash (markup) tili.", false),
		tfQ("`==` va `===` JavaScript'da bir xil ishlaydi.", "Noto'g'ri, `===` tip ham tekshiradi.", false),
	}
	for _, p := range questions {
		p.CategoryID = categoryID
		if _, err := q.CreateQuestion(ctx, p); err != nil {
			log.Fatalf("savol yozish: %v", err)
		}
	}
	log.Printf("✓ programming savollar: %d", len(questions))
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
