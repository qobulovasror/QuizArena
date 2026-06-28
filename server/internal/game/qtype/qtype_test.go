package qtype

import (
	"encoding/json"
	"testing"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		choice  any
		correct any
		want    bool
	}{
		{"mcq to'g'ri", "mcq", map[string]string{"optionId": "b"}, map[string]string{"optionId": "b"}, true},
		{"mcq xato", "mcq", map[string]string{"optionId": "a"}, map[string]string{"optionId": "b"}, false},
		{"mcq bo'sh", "mcq", map[string]string{"optionId": ""}, map[string]string{"optionId": ""}, false},
		{"true_false to'g'ri", "true_false", map[string]bool{"value": true}, map[string]bool{"value": true}, true},
		{"true_false xato", "true_false", map[string]bool{"value": false}, map[string]bool{"value": true}, false},
		{"numeric tolerance ichida", "numeric", map[string]float64{"value": 12.1}, map[string]float64{"value": 12, "tolerance": 0.5}, true},
		{"numeric tolerancedan tashqari", "numeric", map[string]float64{"value": 13}, map[string]float64{"value": 12, "tolerance": 0.5}, false},
		{"type_answer normalize", "type_answer", map[string]string{"text": "  WeNt "}, map[string][]string{"accepted": {"went"}}, true},
		{"type_answer xato", "type_answer", map[string]string{"text": "goed"}, map[string][]string{"accepted": {"went"}}, false},
		{"multi_select tartibsiz", "multi_select", map[string][]string{"optionIds": {"b", "a"}}, map[string][]string{"optionIds": {"a", "b"}}, true},
		{"multi_select kam", "multi_select", map[string][]string{"optionIds": {"a"}}, map[string][]string{"optionIds": {"a", "b"}}, false},

		{"match to'g'ri", "match",
			map[string]map[string]string{"pairs": {"l1": "r2", "l2": "r1"}},
			map[string]map[string]string{"pairs": {"l1": "r2", "l2": "r1"}}, true},
		{"match xato juft", "match",
			map[string]map[string]string{"pairs": {"l1": "r1", "l2": "r2"}},
			map[string]map[string]string{"pairs": {"l1": "r2", "l2": "r1"}}, false},
		{"match to'liqsiz", "match",
			map[string]map[string]string{"pairs": {"l1": "r2"}},
			map[string]map[string]string{"pairs": {"l1": "r2", "l2": "r1"}}, false},

		{"categorize to'g'ri", "categorize",
			map[string]map[string]string{"assign": {"go": "verb", "cat": "noun"}},
			map[string]map[string]string{"assign": {"go": "verb", "cat": "noun"}}, true},
		{"categorize xato", "categorize",
			map[string]map[string]string{"assign": {"go": "noun", "cat": "noun"}},
			map[string]map[string]string{"assign": {"go": "verb", "cat": "noun"}}, false},

		{"ordering to'g'ri", "ordering",
			map[string][]string{"order": {"a", "b", "c"}},
			map[string][]string{"order": {"a", "b", "c"}}, true},
		{"ordering teskari", "ordering",
			map[string][]string{"order": {"c", "b", "a"}},
			map[string][]string{"order": {"a", "b", "c"}}, false},
		{"ordering kalta", "ordering",
			map[string][]string{"order": {"a", "b"}},
			map[string][]string{"order": {"a", "b", "c"}}, false},

		{"cloze to'g'ri (normalize)", "cloze",
			map[string][]string{"blanks": {" Went ", "GONE"}},
			map[string][]any{"blanks": {map[string][]string{"accepted": {"went"}}, map[string][]string{"accepted": {"gone", "goned"}}}}, true},
		{"cloze bitta xato", "cloze",
			map[string][]string{"blanks": {"went", "goed"}},
			map[string][]any{"blanks": {map[string][]string{"accepted": {"went"}}, map[string][]string{"accepted": {"gone"}}}}, false},
		{"cloze bo'sh bo'shliq", "cloze",
			map[string][]string{"blanks": {"went", ""}},
			map[string][]any{"blanks": {map[string][]string{"accepted": {"went"}}, map[string][]string{"accepted": {"gone"}}}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := For(tc.typ).Validate(raw(tc.choice), raw(tc.correct))
			if got != tc.want {
				t.Fatalf("%s: kutilgan %v, keldi %v", tc.name, tc.want, got)
			}
		})
	}
}
