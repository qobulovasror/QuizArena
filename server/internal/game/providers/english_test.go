package providers

import (
	"strings"
	"testing"
)

func TestEnglishVerbGenerates(t *testing.T) {
	p, err := NewEnglishVerb()
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if len(p.verbs) != 110 {
		t.Fatalf("110 fe'l kutilgan, %d", len(p.verbs))
	}

	qs, err := p.Questions(20)
	if err != nil || len(qs) != 20 {
		t.Fatalf("20 savol kutilgan: %v, %d", err, len(qs))
	}

	for _, q := range qs {
		if q.Type != "mcq" || q.Prompt == "" {
			t.Fatalf("yaroqsiz savol: %+v", q)
		}
		if len(q.Options) != 4 {
			t.Fatalf("4 variant kutilgan, %d", len(q.Options))
		}
		// Aniq bitta to'g'ri variant.
		correct := 0
		seen := map[string]bool{}
		for _, o := range q.Options {
			if o.Correct {
				correct++
			}
			low := strings.ToLower(o.Text)
			if seen[low] {
				t.Fatalf("takror variant: %q (%s)", o.Text, q.Prompt)
			}
			seen[low] = true
		}
		if correct != 1 {
			t.Fatalf("aniq 1 to'g'ri kutilgan, %d (%s)", correct, q.Prompt)
		}
	}
}
