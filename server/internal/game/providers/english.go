package providers

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/azizbek12234/quizarena/server/internal/state"
)

//go:embed data/irregularVerb.json
var irregularVerbJSON []byte

type verb struct {
	Word        string `json:"word"`
	PS          string `json:"ps"`
	PP          string `json:"pp"`
	Translation struct {
		Uzbek string `json:"Uzbek"`
	} `json:"translation"`
}

// EnglishVerb — ingliz tili noto'g'ri fe'llari uchun generativ provider.
// Eski generateVerb.js distractor mantig'i Go'ga ko'chirilgan (savollar har gal generatsiya).
type EnglishVerb struct {
	verbs []verb
}

func NewEnglishVerb() (*EnglishVerb, error) {
	var vs []verb
	if err := json.Unmarshal(irregularVerbJSON, &vs); err != nil {
		return nil, fmt.Errorf("irregularVerb.json parse: %w", err)
	}
	if len(vs) < 4 {
		return nil, fmt.Errorf("fe'llar yetarli emas: %d", len(vs))
	}
	return &EnglishVerb{verbs: vs}, nil
}

func (e *EnglishVerb) Questions(count int) ([]state.Question, error) {
	if count <= 0 {
		count = 1
	}
	out := make([]state.Question, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, e.one(i))
	}
	return out, nil
}

func (e *EnglishVerb) one(idx int) state.Question {
	v := e.verbs[rand.Intn(len(e.verbs))]
	askPP := rand.Intn(2) == 0

	var correct, other, label string
	if askPP {
		correct, other, label = v.PP, v.PS, "Past Participle (V3)"
	} else {
		correct, other, label = v.PS, v.PP, "Past Simple (V2)"
	}

	texts := e.options(v, correct, other)
	opts := make([]state.Option, len(texts))
	for i, t := range texts {
		opts[i] = state.Option{ID: fmt.Sprintf("o%d", i), Text: t, Correct: strings.EqualFold(t, correct)}
	}

	return state.Question{
		ID:          fmt.Sprintf("en-%s-%d", strings.ToLower(v.Word), idx),
		Type:        "mcq",
		Prompt:      fmt.Sprintf("«%s» fe'lining %s shakli?", v.Word, label),
		Explanation: fmt.Sprintf("%s → %s → %s — %s", v.Word, v.PS, v.PP, v.Translation.Uzbek),
		Options:     opts,
	}
}

// options — 1 to'g'ri + 3 distractor (jami 4, takrorsiz, katta/kichik harfdan qat'i nazar).
func (e *EnglishVerb) options(v verb, correct, other string) []string {
	seen := map[string]bool{strings.ToLower(strings.TrimSpace(correct)): true}
	out := []string{correct}

	push := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || len(out) >= 4 {
			return
		}
		k := strings.ToLower(s)
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, s)
	}

	push(other)         // ikkinchi shakl (V2/V3)
	push(v.Word)        // asos shakli
	push(v.Word + "ed") // soxta "regular" shakl
	for len(out) < 4 {  // qolganini boshqa fe'llardan to'ldirish
		rv := e.verbs[rand.Intn(len(e.verbs))]
		push(rv.PS)
		push(rv.PP)
	}
	return out
}
