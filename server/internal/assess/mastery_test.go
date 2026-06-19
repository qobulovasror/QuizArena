package assess

import "testing"

func TestUpdateDirection(t *testing.T) {
	if up := Update(50, true); up <= 50 {
		t.Fatalf("to'g'ri javob darajani oshirishi kerak: %v", up)
	}
	if down := Update(50, false); down >= 50 {
		t.Fatalf("xato javob darajani kamaytirishi kerak: %v", down)
	}
}

func TestUpdateConverges(t *testing.T) {
	m := DefaultMastery
	for i := 0; i < 30; i++ {
		m = Update(m, true)
	}
	if m < 95 {
		t.Fatalf("ketma-ket to'g'ri javoblardan keyin ~100 kutilgan: %v", m)
	}
	for i := 0; i < 30; i++ {
		m = Update(m, false)
	}
	if m > 5 {
		t.Fatalf("ketma-ket xatolardan keyin ~0 kutilgan: %v", m)
	}
}

func TestUpdateBounds(t *testing.T) {
	if Update(100, true) > 100 || Update(0, false) < 0 {
		t.Fatal("mastery 0..100 oralig'ida qolishi kerak")
	}
}
