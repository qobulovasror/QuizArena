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
