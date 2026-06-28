package httpapi

import (
	"encoding/json"
	"sort"
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

// shuffleOptions hech qanday elementni yo'qotmaydi/qo'shmaydi; edge case'lar xavfsiz.
func TestShuffleOptions(t *testing.T) {
	if shuffleOptions(nil) != nil {
		t.Fatal("nil → nil kutilgan")
	}
	one := []byte(`[{"id":"a"}]`)
	if string(shuffleOptions(one)) != string(one) {
		t.Fatal("bitta element o'zgarmasligi kerak")
	}
	multi := []byte(`[{"id":"a"},{"id":"b"},{"id":"c"}]`)
	out := shuffleOptions(multi)
	var got []map[string]string
	if err := json.Unmarshal(out, &got); err != nil || len(got) != 3 {
		t.Fatalf("3 element saqlanishi kerak: %s", out)
	}
	ids := []string{got[0]["id"], got[1]["id"], got[2]["id"]}
	sort.Strings(ids)
	if ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Fatalf("aynan a,b,c id'lari kutilgan: %v", ids)
	}
}
