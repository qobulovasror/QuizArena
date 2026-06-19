package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/assess"
	"github.com/azizbek12234/quizarena/server/internal/game/qtype"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

type assessHandler struct {
	q      *store.Queries
	logger *slog.Logger
}

// Test savoli — to'g'ri javob YO'Q (baholash, server hisoblaydi).
type assessQ struct {
	QuestionID string          `json:"questionId"`
	Type       string          `json:"type"`
	Prompt     string          `json:"prompt"`
	Options    json.RawMessage `json:"options,omitempty"`
}

// questions — soha bo'yicha test savollari (javobsiz).
func (h *assessHandler) questions(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("subject")
	if slug == "" {
		slug = "general"
	}
	subj, err := h.q.GetSubjectBySlug(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "soha topilmadi")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	count := 5
	if c, _ := strconv.Atoi(r.URL.Query().Get("count")); c >= 1 && c <= 20 {
		count = c
	}
	rows, err := h.q.RandomQuestionsBySubject(r.Context(), store.RandomQuestionsBySubjectParams{
		SubjectID: subj.ID, Limit: int32(count),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "savollarni olishda xato")
		return
	}
	out := make([]assessQ, 0, len(rows))
	for _, q := range rows {
		out = append(out, assessQ{QuestionID: q.ID.String(), Type: q.Type, Prompt: q.Prompt, Options: q.Options})
	}
	writeJSON(w, http.StatusOK, out)
}

type submitReq struct {
	Answers []struct {
		QuestionID string          `json:"questionId"`
		Choice     json.RawMessage `json:"choice"`
	} `json:"answers"`
}

type submitResult struct {
	Correct int              `json:"correct"`
	Total   int              `json:"total"`
	Results []map[string]any `json:"results"`
}

// submit — javoblarni server-side baholaydi va mastery'ni yangilaydi.
func (h *assessHandler) submit(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) == 0 {
		writeErr(w, http.StatusBadRequest, "javoblar yo'q")
		return
	}

	type mst struct {
		mastery  float64
		attempts int32
	}
	local := map[uuid.UUID]mst{}
	correct := 0
	results := make([]map[string]any, 0, len(req.Answers))

	for _, a := range req.Answers {
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil {
			continue
		}
		q, err := h.q.GetQuestionByID(r.Context(), qid)
		if err != nil {
			continue
		}
		ok := qtype.For(q.Type).Validate(a.Choice, q.Correct)
		if ok {
			correct++
		}
		results = append(results, map[string]any{"questionId": a.QuestionID, "correct": ok})

		cur, found := local[q.CategoryID]
		if !found {
			cur = mst{mastery: assess.DefaultMastery}
			if dbm, err := h.q.GetMastery(r.Context(), store.GetMasteryParams{UserID: uid, CategoryID: q.CategoryID}); err == nil {
				cur = mst{mastery: dbm.Mastery, attempts: dbm.Attempts}
			}
		}
		cur.mastery = assess.Update(cur.mastery, ok)
		cur.attempts++
		local[q.CategoryID] = cur
	}

	for cat, m := range local {
		if err := h.q.UpsertMastery(r.Context(), store.UpsertMasteryParams{
			UserID: uid, CategoryID: cat, Mastery: m.mastery, Attempts: m.attempts,
		}); err != nil {
			h.logger.Error("mastery upsert", "err", err)
		}
	}
	writeJSON(w, http.StatusOK, submitResult{Correct: correct, Total: len(req.Answers), Results: results})
}

type masteryDTO struct {
	Subject  string  `json:"subject"`
	Category string  `json:"category"`
	Mastery  float64 `json:"mastery"`
	Attempts int32   `json:"attempts"`
}

// mastery — foydalanuvchining soha/kategoriya darajalari.
func (h *assessHandler) mastery(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
	rows, err := h.q.ListMasteryByUser(r.Context(), uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	out := make([]masteryDTO, 0, len(rows))
	for _, m := range rows {
		out = append(out, masteryDTO{Subject: m.SubjectName, Category: m.CategoryName, Mastery: m.Mastery, Attempts: m.Attempts})
	}
	writeJSON(w, http.StatusOK, out)
}
