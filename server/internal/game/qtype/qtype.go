// Package qtype — savol turi strategiyalari (server-authoritative baholash).
//
// Har tur faqat Validate'da farq qiladi (Render/Reveal generic — Options va Correct
// to'g'ridan-to'g'ri ishlatiladi). Yangi tur qo'shish = bitta strategiya + For() ga qator.
package qtype

import (
	"encoding/json"
	"math"
	"strings"
)

// QuestionType — xom javobni (choice) to'g'ri javob (correct) bilan solishtiradi.
type QuestionType interface {
	Validate(choice, correct json.RawMessage) bool
}

// For — tur nomi bo'yicha strategiya (noma'lum → mcq).
func For(t string) QuestionType {
	switch t {
	case "true_false":
		return trueFalse{}
	case "numeric":
		return numeric{}
	case "type_answer", "fill_blank":
		return typeAnswer{}
	case "multi_select":
		return multiSelect{}
	default:
		return mcq{}
	}
}

type mcq struct{}

func (mcq) Validate(choice, correct json.RawMessage) bool {
	var c, k struct {
		OptionID string `json:"optionId"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return c.OptionID != "" && c.OptionID == k.OptionID
}

type trueFalse struct{}

func (trueFalse) Validate(choice, correct json.RawMessage) bool {
	var c struct {
		Value *bool `json:"value"`
	}
	var k struct {
		Value bool `json:"value"`
	}
	if json.Unmarshal(choice, &c) != nil || c.Value == nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return *c.Value == k.Value
}

type numeric struct{}

func (numeric) Validate(choice, correct json.RawMessage) bool {
	var c struct {
		Value *float64 `json:"value"`
	}
	var k struct {
		Value     float64 `json:"value"`
		Tolerance float64 `json:"tolerance"`
	}
	if json.Unmarshal(choice, &c) != nil || c.Value == nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return math.Abs(*c.Value-k.Value) <= k.Tolerance
}

type typeAnswer struct{}

func (typeAnswer) Validate(choice, correct json.RawMessage) bool {
	var c struct {
		Text string `json:"text"`
	}
	var k struct {
		Accepted []string `json:"accepted"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	ct := norm(c.Text)
	if ct == "" {
		return false
	}
	for _, a := range k.Accepted {
		if norm(a) == ct {
			return true
		}
	}
	return false
}

type multiSelect struct{}

func (multiSelect) Validate(choice, correct json.RawMessage) bool {
	var c, k struct {
		OptionIDs []string `json:"optionIds"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return sameSet(c.OptionIDs, k.OptionIDs)
}

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func sameSet(a, b []string) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}
	m := make(map[string]int, len(a))
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		if m[x] == 0 {
			return false
		}
		m[x]--
	}
	return true
}
