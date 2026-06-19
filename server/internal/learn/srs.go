// Package learn — Spaced Repetition (SM-2) mantiqi.
package learn

import (
	"math"
	"time"
)

// Card — SM-2 holati (DB'dagi srs_cards yozuvi).
type Card struct {
	Ease     float64
	Interval int // kun
	Reps     int
}

// NewCard — birinchi marta ko'rilgan savol uchun boshlang'ich holat.
func NewCard() Card { return Card{Ease: 2.5, Interval: 0, Reps: 0} }

// Grade — foydalanuvchining o'z bahosi (sodda 3 daraja).
const (
	GradeForgot = 0 // qaytadan
	GradeGood   = 1 // yaxshi
	GradeEasy   = 2 // oson
)

// quality — sodda bahoni SM-2 sifat (0..5) ga o'giradi.
var quality = map[int]int{GradeForgot: 2, GradeGood: 4, GradeEasy: 5}

// Review — SM-2: yangi holat + keyingi takror sanasini qaytaradi.
func Review(c Card, grade int, now time.Time) (Card, time.Time) {
	q, ok := quality[grade]
	if !ok {
		q = GradeGood
	}

	if q < 3 { // unutgan — boshdan
		c.Reps = 0
		c.Interval = 1
	} else {
		switch c.Reps {
		case 0:
			c.Interval = 1
		case 1:
			c.Interval = 6
		default:
			c.Interval = int(math.Round(float64(c.Interval) * c.Ease))
		}
		c.Reps++
	}

	// Ease yangilash (SM-2 formulasi), pastki chegara 1.3.
	c.Ease += 0.1 - float64(5-q)*(0.08+float64(5-q)*0.02)
	if c.Ease < 1.3 {
		c.Ease = 1.3
	}

	due := now.AddDate(0, 0, c.Interval)
	return c, due
}
