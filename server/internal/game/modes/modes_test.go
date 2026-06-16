package modes

import (
	"testing"

	"github.com/azizbek12234/quizarena/server/internal/state"
)

func liveQ() *state.LiveQuestion {
	return &state.LiveQuestion{AskedAt: 0, Deadline: 1000}
}

func TestClassicScoring(t *testing.T) {
	p := &state.Player{}
	Classic{}.OnAnswer(p, liveQ(), true, 0) // darhol to'g'ri
	if p.Score < 199 || p.CorrectCnt != 1 {
		t.Fatalf("classic to'g'ri: ~200 ball, 1 to'g'ri kutilgan: %+v", p)
	}
	Classic{}.OnAnswer(p, liveQ(), false, 0) // xato
	if p.CorrectCnt != 1 || p.Eliminated {
		t.Fatalf("classic xato o'zgartirmasligi kerak: %+v", p)
	}
}

func TestSurvivalEliminate(t *testing.T) {
	p := &state.Player{}
	Survival{}.OnAnswer(p, liveQ(), false, 0)
	if !p.Eliminated {
		t.Fatal("xato javob survival'da eliminate qilishi kerak")
	}
	ok := &state.Player{}
	Survival{}.OnAnswer(ok, liveQ(), true, 0)
	if ok.Eliminated || ok.CorrectCnt != 1 {
		t.Fatalf("to'g'ri javob tirik qoldirishi kerak: %+v", ok)
	}
}

func TestSurvivalEndEarly(t *testing.T) {
	r := &state.Room{Players: map[string]*state.Player{
		"a": {Eliminated: false},
		"b": {Eliminated: true},
	}}
	if !(Survival{}).EndEarly(r) {
		t.Fatal("1 tirik qolganda tugashi kerak")
	}
	r.Players["b"].Eliminated = false
	if (Survival{}).EndEarly(r) {
		t.Fatal("2 tirik bo'lganda davom etishi kerak")
	}
}
