// Package providers — savol manbalari (statik bank yoki generativ).
//
// Bosqich 1 da Sample (qattiq kodlangan) — engine'ni DB'siz sinash uchun.
// Bosqich 1+ da EnglishVerbProvider (irregularVerb.json) + DB seed qo'shiladi.
package providers

import "github.com/azizbek12234/quizarena/server/internal/state"

// Sample — vaqtinchalik namuna provider (bir nechta mcq savol).
type Sample struct{}

func NewSample() Sample { return Sample{} }

var sampleBank = []state.Question{
	{ID: "s1", Type: "mcq", Prompt: "2 + 2 = ?", Explanation: "Oddiy qo'shish.",
		Options: []state.Option{{ID: "a", Text: "3"}, {ID: "b", Text: "4", Correct: true}, {ID: "c", Text: "5"}, {ID: "d", Text: "22"}}},
	{ID: "s2", Type: "mcq", Prompt: "Past simple of 'go'?", Explanation: "go → went → gone.",
		Options: []state.Option{{ID: "a", Text: "goed"}, {ID: "b", Text: "gone"}, {ID: "c", Text: "went", Correct: true}, {ID: "d", Text: "going"}}},
	{ID: "s3", Type: "mcq", Prompt: "O'zbekiston poytaxti?", Explanation: "Toshkent.",
		Options: []state.Option{{ID: "a", Text: "Samarqand"}, {ID: "b", Text: "Toshkent", Correct: true}, {ID: "c", Text: "Buxoro"}, {ID: "d", Text: "Xiva"}}},
	{ID: "s4", Type: "mcq", Prompt: "HTTP qaysi portda (standart)?", Explanation: "80.",
		Options: []state.Option{{ID: "a", Text: "21"}, {ID: "b", Text: "443"}, {ID: "c", Text: "80", Correct: true}, {ID: "d", Text: "8080"}}},
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
