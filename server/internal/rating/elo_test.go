package rating

import "testing"

func TestExpected(t *testing.T) {
	if e := Expected(1000, 1000); e != 0.5 {
		t.Fatalf("teng reytingda 0.5 kutilgan: %v", e)
	}
	if Expected(1200, 1000) <= 0.5 {
		t.Fatal("kuchli o'yinchi uchun > 0.5 kutilgan")
	}
}

func TestNextWinLossDraw(t *testing.T) {
	// Teng o'yinchilar (exp=0.5, K=32): g'olib +16, mag'lub -16, durang 0.
	if g := Next(1000, 1000, 1); g != 1016 {
		t.Fatalf("g'alaba 1016 kutilgan: %d", g)
	}
	if l := Next(1000, 1000, 0); l != 984 {
		t.Fatalf("mag'lubiyat 984 kutilgan: %d", l)
	}
	if d := Next(1000, 1000, 0.5); d != 1000 {
		t.Fatalf("durang 1000 kutilgan: %d", d)
	}
}

func TestUnderdogGainsMore(t *testing.T) {
	weakGain := Next(1000, 1400, 1) - 1000
	strongGain := Next(1400, 1000, 1) - 1400
	if weakGain <= strongGain {
		t.Fatalf("kuchsiz g'olib ko'proq olishi kerak: weak=%d strong=%d", weakGain, strongGain)
	}
}
