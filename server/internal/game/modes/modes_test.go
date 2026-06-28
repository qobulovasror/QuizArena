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

func TestTimeAttackScoring(t *testing.T) {
	p := &state.Player{}
	// Self-paced: tezlik bonusi yo'q — har to'g'ri javob aniq +100.
	TimeAttack{}.OnAnswer(p, liveQ(), true, 1000) // deadline'da bo'lsa ham to'liq 100
	TimeAttack{}.OnAnswer(p, liveQ(), true, 0)
	if p.Score != 200 || p.CorrectCnt != 2 {
		t.Fatalf("time_attack: 200 ball, 2 to'g'ri kutilgan: %+v", p)
	}
	TimeAttack{}.OnAnswer(p, liveQ(), false, 0)
	if p.CorrectCnt != 2 || p.Eliminated {
		t.Fatalf("time_attack xato o'zgartirmasligi kerak: %+v", p)
	}
}

func TestTeamScoring(t *testing.T) {
	p := &state.Player{Team: "A"}
	Team{}.OnAnswer(p, liveQ(), true, 0) // classic kabi: tezlik bonusi bilan ~200
	if p.Score < 199 || p.CorrectCnt != 1 {
		t.Fatalf("team to'g'ri: ~200 ball, 1 to'g'ri kutilgan: %+v", p)
	}
	if (Team{}).EndEarly(&state.Room{}) {
		t.Fatal("team erta tugamasligi kerak")
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
