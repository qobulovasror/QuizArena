package httpapi

import (
	"testing"
	"time"
)

func TestTournamentStatus(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	starts := now.Add(-time.Hour)
	ends := now.Add(time.Hour)

	cases := []struct {
		name   string
		now    time.Time
		starts time.Time
		ends   time.Time
		want   string
	}{
		{"boshlanmagan", now, now.Add(time.Hour), now.Add(2 * time.Hour), "upcoming"},
		{"faol", now, starts, ends, "active"},
		{"tugagan", now, now.Add(-2 * time.Hour), now.Add(-time.Hour), "finished"},
	}
	for _, c := range cases {
		if got := status(c.now, c.starts, c.ends); got != c.want {
			t.Errorf("%s: status = %q, kutilgan %q", c.name, got, c.want)
		}
	}
}
