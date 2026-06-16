package providers

import (
	"encoding/json"
	"testing"

	"github.com/azizbek12234/quizarena/server/internal/game/qtype"
)

func TestMathSelfConsistent(t *testing.T) {
	qs, err := NewMath().Questions(30)
	if err != nil || len(qs) != 30 {
		t.Fatalf("30 savol kutilgan: %v, %d", err, len(qs))
	}
	for _, q := range qs {
		if q.Type != "numeric" || q.Prompt == "" {
			t.Fatalf("numeric savol kutilgan: %+v", q)
		}
		// To'g'ri qiymatni javob qilib yuborsak, qtype.numeric uni tasdiqlashi kerak.
		var c struct {
			Value float64 `json:"value"`
		}
		if err := json.Unmarshal(q.Correct, &c); err != nil {
			t.Fatalf("Correct parse: %v (%s)", err, q.Prompt)
		}
		choice, _ := json.Marshal(map[string]float64{"value": c.Value})
		if !qtype.For("numeric").Validate(choice, q.Correct) {
			t.Fatalf("math savol o'z javobini tasdiqlamadi: %s (=%v)", q.Prompt, c.Value)
		}
	}
}
