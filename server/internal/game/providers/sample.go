// Package providers — savol manbalari (statik bank yoki generativ).
//
// Bosqich 1 da Sample (qattiq kodlangan) — engine'ni DB'siz sinash uchun.
// Bosqich 1+ da EnglishVerbProvider (irregularVerb.json) + DB seed qo'shiladi.
package providers

import (
	"encoding/json"

	"github.com/azizbek12234/quizarena/server/internal/state"
)

// Sample — vaqtinchalik namuna provider (bir nechta mcq savol).
type Sample struct{}

func NewSample() Sample { return Sample{} }

// mcq — variantli savol quradi; correctIdx — to'g'ri variant indeksi.
func mcq(id, prompt, explanation string, correctIdx int, texts ...string) state.Question {
	ids := []string{"a", "b", "c", "d", "e"}
	options := make([]state.Option, len(texts))
	for i, t := range texts {
		options[i] = state.Option{ID: ids[i], Text: t}
	}
	correct, _ := json.Marshal(map[string]string{"optionId": ids[correctIdx]})
	return state.Question{
		ID: id, Type: "mcq", Prompt: prompt, Explanation: explanation,
		Options: options, Correct: correct,
	}
}

var sampleBank = []state.Question{
	mcq("s1", "2 + 2 = ?", "Oddiy qo'shish.", 1, "3", "4", "5", "22"),
	mcq("s2", "Past simple of 'go'?", "go → went → gone.", 2, "goed", "gone", "went", "going"),
	mcq("s3", "O'zbekiston poytaxti?", "Toshkent.", 1, "Samarqand", "Toshkent", "Buxoro", "Xiva"),
	mcq("s4", "HTTP standart porti?", "80.", 2, "21", "443", "80", "8080"),
}

// Questions — so'ralgan sondagi savol qaytaradi (bank kichik bo'lsa aylantiradi).
func (Sample) Questions(count int) ([]state.Question, error) {
	if count <= 0 {
		count = 1
	}
	out := make([]state.Question, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, sampleBank[i%len(sampleBank)])
	}
	return out, nil
}
