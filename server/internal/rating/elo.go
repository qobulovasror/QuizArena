// Package rating — ELO reyting hisoblash (1v1 duel uchun, subject bo'yicha). PLAN §4.
package rating

import "math"

const (
	Default = 1000 // boshlang'ich reyting
	K       = 32   // K-faktor (bitta o'yindagi maksimal o'zgarish kuchi)
)

// Expected — a o'yinchining b ga qarshi kutilgan natijasi (0..1).
func Expected(a, b int) float64 {
	return 1 / (1 + math.Pow(10, float64(b-a)/400))
}

// Next — yangi reyting. score: 1=g'alaba, 0.5=durang, 0=mag'lubiyat.
func Next(rating, opponent int, score float64) int {
	return rating + int(math.Round(K*(score-Expected(rating, opponent))))
}
