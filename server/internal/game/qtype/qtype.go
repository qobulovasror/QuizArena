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
	case "match":
		return match{}
	case "categorize":
		return categorize{}
	case "ordering":
		return ordering{}
	case "cloze":
		return cloze{}
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

// match — juftlash: leftId→rightId moslik to'liq mos kelishi kerak.
type match struct{}

func (match) Validate(choice, correct json.RawMessage) bool {
	var c, k struct {
		Pairs map[string]string `json:"pairs"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return sameMapping(c.Pairs, k.Pairs)
}

// categorize — har element to'g'ri toifaga (itemId→catId).
type categorize struct{}

func (categorize) Validate(choice, correct json.RawMessage) bool {
	var c, k struct {
		Assign map[string]string `json:"assign"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	return sameMapping(c.Assign, k.Assign)
}

// ordering — id'lar ketma-ketligi aniq mos bo'lishi kerak.
type ordering struct{}

func (ordering) Validate(choice, correct json.RawMessage) bool {
	var c, k struct {
		Order []string `json:"order"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	if len(c.Order) == 0 || len(c.Order) != len(k.Order) {
		return false
	}
	for i := range k.Order {
		if c.Order[i] != k.Order[i] {
			return false
		}
	}
	return true
}

// cloze — ko'p bo'sh joy; har bo'shliq mos accepted ro'yxatida bo'lishi kerak.
type cloze struct{}

func (cloze) Validate(choice, correct json.RawMessage) bool {
	var c struct {
		Blanks []string `json:"blanks"`
	}
	var k struct {
		Blanks []struct {
			Accepted []string `json:"accepted"`
		} `json:"blanks"`
	}
	if json.Unmarshal(choice, &c) != nil || json.Unmarshal(correct, &k) != nil {
		return false
	}
	if len(c.Blanks) == 0 || len(c.Blanks) != len(k.Blanks) {
		return false
	}
	for i, blank := range k.Blanks {
		got := norm(c.Blanks[i])
		if got == "" {
			return false
		}
		matched := false
		for _, a := range blank.Accepted {
			if norm(a) == got {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// sameMapping — ikki string→string xarita aynan teng (bo'sh bo'lmasa).
func sameMapping(a, b map[string]string) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}
	for k, v := range b {
		if a[k] != v {
			return false
		}
	}
	return true
}

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
