// Package assess — bilim darajasi (mastery) hisobi.
package assess

// alpha — eksponensial silliqlash koeffitsienti (yangi javobning og'irligi).
const alpha = 0.25

// DefaultMastery — yangi kategoriya uchun boshlang'ich daraja.
const DefaultMastery = 50.0

// Update — javob natijasiga qarab mastery'ni yangilaydi (EMA, 0..100).
//
//	to'g'ri  → 100 tomon siljiydi
//	xato     → 0 tomon siljiydi
func Update(current float64, correct bool) float64 {
	target := 0.0
	if correct {
		target = 100.0
	}
	m := current + alpha*(target-current)
	if m < 0 {
		m = 0
	}
	if m > 100 {
		m = 100
	}
	return m
}
