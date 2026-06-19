package learn

import (
	"testing"
	"time"
)

func TestReviewProgression(t *testing.T) {
	now := time.Now()
	c := NewCard()

	// Ketma-ket "yaxshi": interval 1 → 6 → o'sib boradi.
	c, due1 := Review(c, GradeGood, now)
	if c.Interval != 1 || c.Reps != 1 {
		t.Fatalf("1-takror: interval 1, reps 1 kutilgan: %+v", c)
	}
	c, _ = Review(c, GradeGood, now)
	if c.Interval != 6 || c.Reps != 2 {
		t.Fatalf("2-takror: interval 6 kutilgan: %+v", c)
	}
	c, due3 := Review(c, GradeGood, now)
	if c.Interval <= 6 {
		t.Fatalf("3-takror: interval > 6 kutilgan: %+v", c)
	}
	if !due3.After(due1) {
		t.Fatal("keyingi takror sanasi uzoqroq bo'lishi kerak")
	}
}

func TestForgotResets(t *testing.T) {
	now := time.Now()
	c := Card{Ease: 2.6, Interval: 20, Reps: 5}
	c, due := Review(c, GradeForgot, now)
	if c.Reps != 0 || c.Interval != 1 {
		t.Fatalf("unutganda reps=0, interval=1 kutilgan: %+v", c)
	}
	if c.Ease >= 2.6 {
		t.Fatalf("unutganda ease kamayishi kerak: %v", c.Ease)
	}
	if due.Before(now) {
		t.Fatal("due kelajakda bo'lishi kerak")
	}
}

func TestEaseFloor(t *testing.T) {
	now := time.Now()
	c := Card{Ease: 1.3, Interval: 1, Reps: 1}
	for i := 0; i < 5; i++ {
		c, _ = Review(c, GradeForgot, now)
	}
	if c.Ease < 1.3 {
		t.Fatalf("ease 1.3 dan past tushmasligi kerak: %v", c.Ease)
	}
}
