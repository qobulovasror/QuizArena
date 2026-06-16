package providers

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/azizbek12234/quizarena/server/internal/state"
)

// Math — matematika uchun generativ provider (arifmetika + foiz).
// `numeric` turidan foydalanadi (qtype.numeric baholaydi) — QuestionType abstraksiyasini namoyish etadi.
type Math struct{}

func NewMath() Math { return Math{} }

func (Math) Questions(count int) ([]state.Question, error) {
	if count <= 0 {
		count = 1
	}
	out := make([]state.Question, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, genMath(i))
	}
	return out, nil
}

func genMath(idx int) state.Question {
	switch rand.Intn(4) {
	case 0: // ko'paytirish
		a, b := rand.Intn(11)+2, rand.Intn(11)+2
		return numericQ(idx, fmt.Sprintf("%d × %d = ?", a, b), float64(a*b))
	case 1: // qo'shish
		a, b := rand.Intn(90)+10, rand.Intn(90)+10
		return numericQ(idx, fmt.Sprintf("%d + %d = ?", a, b), float64(a+b))
	case 2: // ayirish (manfiy bo'lmasin)
		a, b := rand.Intn(90)+10, rand.Intn(90)+10
		if b > a {
			a, b = b, a
		}
		return numericQ(idx, fmt.Sprintf("%d − %d = ?", a, b), float64(a-b))
	default: // foiz
		pct := []int{10, 20, 25, 50}[rand.Intn(4)]
		base := (rand.Intn(18) + 2) * 10
		return numericQ(idx, fmt.Sprintf("%d ning %d%% qancha?", base, pct), float64(base*pct)/100)
	}
}

func numericQ(idx int, prompt string, answer float64) state.Question {
	correct, _ := json.Marshal(map[string]float64{"value": answer, "tolerance": 0})
	return state.Question{
		ID:      fmt.Sprintf("math-%d", idx),
		Type:    "numeric",
		Prompt:  prompt,
		Correct: correct,
	}
}
