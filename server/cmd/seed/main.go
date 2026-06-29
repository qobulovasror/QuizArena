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

	// Yangi savol formatlari (ordering/cloze/match/categorize) — namuna bank.
	tcat := ensureCategory(gen.ID, "formats", "Turli formatlar")
	seedTypes(tcat.ID)

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

// seedTypes — har yangi turdan bittadan namuna (ordering/cloze/match/categorize/anagram/type_answer).
func seedTypes(categoryID uuid.UUID) {
	n, err := q.CountQuestionsByCategory(ctx, categoryID)
	if err != nil {
		log.Fatalf("savol soni: %v", err)
	}
	if n > 0 {
		log.Printf("• format savollar mavjud (%d)", n)
		return
	}
	questions := []store.CreateQuestionParams{
		orderingQ("So'zlardan to'g'ri jumla tuzing:", "To'g'ri tartib: I like tea.",
			[][2]string{{"o1", "I"}, {"o2", "like"}, {"o3", "tea"}}, []string{"o1", "o2", "o3"}),
		clozeQ("Bo'sh joylarni to'ldiring: 2 + 2 = ___ , 10 − 4 = ___", "4 va 6.",
			[][]string{{"4"}, {"6"}}),
		matchQ("So'zni tarjimasiga moslang:", "cat → mushuk, dog → it.",
			[][2]string{{"l1", "cat"}, {"l2", "dog"}}, [][2]string{{"r1", "mushuk"}, {"r2", "it"}},
			map[string]string{"l1": "r1", "l2": "r2"}),
		categorizeQ("So'zlarni turkumga ajrating:", "olma/nok — meva; mushuk — hayvon.",
			[][2]string{{"i1", "olma"}, {"i2", "mushuk"}, {"i3", "nok"}},
			[][2]string{{"c1", "Meva"}, {"c2", "Hayvon"}},
			map[string]string{"i1": "c1", "i2": "c2", "i3": "c1"}),
		anagramQ("Harflardan so'z tuz: T-E-N-W", "Aralash harflar: went.", []string{"went"}),
		typeAnswerQ("«go» fe'lining 2-shakli?", "go → went.", []string{"went"}),
	}
	for _, p := range questions {
		p.CategoryID = categoryID
		if _, err := q.CreateQuestion(ctx, p); err != nil {
			log.Fatalf("savol yozish: %v", err)
		}
	}
	log.Printf("✓ format savollar: %d", len(questions))
}

// optsFrom — [id,text] juftliklarini {id,text} obyektlariga (state.Option JSON).
func optsFrom(pairs [][2]string) []map[string]string {
	out := make([]map[string]string, len(pairs))
	for i, p := range pairs {
		out[i] = map[string]string{"id": p[0], "text": p[1]}
	}
	return out
}

func orderingQ(prompt, expl string, items [][2]string, order []string) store.CreateQuestionParams {
	ob, _ := json.Marshal(optsFrom(items))
	cb, _ := json.Marshal(map[string][]string{"order": order})
	e := expl
	return store.CreateQuestionParams{Type: "ordering", Prompt: prompt, Options: ob, Correct: cb, Explanation: &e, Difficulty: 1}
}

func clozeQ(prompt, expl string, accepted [][]string) store.CreateQuestionParams {
	blanks := make([]map[string][]string, len(accepted))
	for i, a := range accepted {
		blanks[i] = map[string][]string{"accepted": a}
	}
	cb, _ := json.Marshal(map[string]any{"blanks": blanks})
	e := expl
	return store.CreateQuestionParams{Type: "cloze", Prompt: prompt, Correct: cb, Explanation: &e, Difficulty: 1}
}

func matchQ(prompt, expl string, left, right [][2]string, pairs map[string]string) store.CreateQuestionParams {
	ob, _ := json.Marshal(optsFrom(left))
	cb, _ := json.Marshal(map[string]map[string]string{"pairs": pairs})
	mb, _ := json.Marshal(map[string]any{"targets": optsFrom(right)})
	e := expl
	return store.CreateQuestionParams{Type: "match", Prompt: prompt, Options: ob, Correct: cb, Meta: mb, Explanation: &e, Difficulty: 1}
}

func categorizeQ(prompt, expl string, items, cats [][2]string, assign map[string]string) store.CreateQuestionParams {
	ob, _ := json.Marshal(optsFrom(items))
	cb, _ := json.Marshal(map[string]map[string]string{"assign": assign})
	mb, _ := json.Marshal(map[string]any{"targets": optsFrom(cats)})
	e := expl
	return store.CreateQuestionParams{Type: "categorize", Prompt: prompt, Options: ob, Correct: cb, Meta: mb, Explanation: &e, Difficulty: 1}
}

// typeAnswerQ — matnli javob: correct = {accepted:[...]}; options yo'q.
func typeAnswerQ(prompt, expl string, accepted []string) store.CreateQuestionParams {
	cb, _ := json.Marshal(map[string][]string{"accepted": accepted})
	e := expl
	return store.CreateQuestionParams{Type: "type_answer", Prompt: prompt, Correct: cb, Explanation: &e, Difficulty: 1}
}

// anagramQ — aralash harflar prompt'da; baholanishi type_answer kabi (accepted ro'yxati).
func anagramQ(prompt, expl string, accepted []string) store.CreateQuestionParams {
	cb, _ := json.Marshal(map[string][]string{"accepted": accepted})
	e := expl
	return store.CreateQuestionParams{Type: "anagram", Prompt: prompt, Correct: cb, Explanation: &e, Difficulty: 1}
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
