// Package modes — o'yin metodi strategiyalari (classic, survival, ...).
//
// Engine umumiy savol-oqimini yuritadi; mode faqat ball/holat farqini belgilaydi.
// Yangi metod = bitta Mode implementatsiyasi + For() ga qator (engine o'zgarmaydi).
package modes

import "github.com/azizbek12234/quizarena/server/internal/state"

type Mode interface {
	Name() string
	// OnAnswer — javob qayta ishlanganda (ball/eliminate). Room locki ushlab turilgan holatda chaqiriladi.
	OnAnswer(p *state.Player, q *state.LiveQuestion, correct bool, now int64)
	// EndEarly — joriy savol tugagach o'yin erta tugashi kerakmi (room locki YO'Q).
	EndEarly(r *state.Room) bool
}

// For — metod nomi bo'yicha strategiya (noma'lum → classic).
func For(name string) Mode {
	switch name {
	case "survival":
		return Survival{}
	default:
		return Classic{}
	}
}

// scoreFor — ball: 100 baza + tezlik bonusi (0..100). Faqat to'g'ri javobga.
func scoreFor(q *state.LiveQuestion, now int64) float64 {
	const base = 100.0
	total := float64(q.Deadline - q.AskedAt)
	if total <= 0 {
		return base
	}
	remaining := float64(q.Deadline - now)
	if remaining < 0 {
		remaining = 0
	}
	return base + base*(remaining/total)
}

// Classic — hamma bir xil savolga javob beradi; tezlik + to'g'rilik = ball.
type Classic struct{}

func (Classic) Name() string { return "classic" }

func (Classic) OnAnswer(p *state.Player, q *state.LiveQuestion, correct bool, now int64) {
	if correct {
		p.Score += scoreFor(q, now)
		p.CorrectCnt++
	}
}

func (Classic) EndEarly(*state.Room) bool { return false }

// Survival — xato javob = o'yindan chiqish. Bittadan kam tirik qolganda tugaydi.
type Survival struct{}

func (Survival) Name() string { return "survival" }

func (Survival) OnAnswer(p *state.Player, q *state.LiveQuestion, correct bool, now int64) {
	if correct {
		p.Score += scoreFor(q, now)
		p.CorrectCnt++
	} else {
		p.Eliminated = true
	}
}

func (Survival) EndEarly(r *state.Room) bool {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	alive := 0
	for _, p := range r.Players {
		if !p.Eliminated {
			alive++
		}
	}
	return alive <= 1
}
