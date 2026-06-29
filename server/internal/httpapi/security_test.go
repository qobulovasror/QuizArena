package httpapi

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	const ip = "1.2.3.4"

	for i := 0; i < 3; i++ {
		if !rl.allow(ip) {
			t.Fatalf("dastlabki %d so'rov ruxsat etilishi kerak", i+1)
		}
	}
	if rl.allow(ip) {
		t.Fatal("limitdan oshgan so'rov rad etilishi kerak")
	}
	// Boshqa IP alohida hisoblanadi.
	if !rl.allow("5.6.7.8") {
		t.Fatal("boshqa IP cheklanmasligi kerak")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := newRateLimiter(1, 20*time.Millisecond)
	const ip = "9.9.9.9"
	if !rl.allow(ip) {
		t.Fatal("birinchi so'rov ruxsat")
	}
	if rl.allow(ip) {
		t.Fatal("oyna ichida ikkinchi so'rov rad etiladi")
	}
	time.Sleep(30 * time.Millisecond)
	if !rl.allow(ip) {
		t.Fatal("oyna tugagach yana ruxsat etilishi kerak")
	}
}
