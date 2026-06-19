package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

type adminHandler struct {
	q        *store.Queries
	validate *validator.Validate
	logger   *slog.Logger
}

// --- Soha ---

type createSubjectReq struct {
	Slug string `json:"slug" validate:"required"`
	Name string `json:"name" validate:"required"`
	Icon string `json:"icon"`
}

func (h *adminHandler) createSubject(w http.ResponseWriter, r *http.Request) {
	var req createSubjectReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	var icon *string
	if req.Icon != "" {
		icon = &req.Icon
	}
	s, err := h.q.CreateSubject(r.Context(), store.CreateSubjectParams{Slug: req.Slug, Name: req.Name, Icon: icon})
	if err != nil {
		writeErr(w, http.StatusBadRequest, "soha yaratib bo'lmadi (slug band?)")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": s.ID.String()})
}

// --- Kategoriya ---

type createCategoryReq struct {
	SubjectID string `json:"subjectId" validate:"required,uuid"`
	Slug      string `json:"slug" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

func (h *adminHandler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	sid, _ := uuid.Parse(req.SubjectID)
	c, err := h.q.CreateCategory(r.Context(), store.CreateCategoryParams{SubjectID: sid, Slug: req.Slug, Name: req.Name})
	if err != nil {
		writeErr(w, http.StatusBadRequest, "kategoriya yaratib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": c.ID.String()})
}

// --- Savol ---

type createQuestionReq struct {
	CategoryID  string          `json:"categoryId" validate:"required,uuid"`
	Type        string          `json:"type" validate:"required"`
	Prompt      string          `json:"prompt" validate:"required"`
	Options     json.RawMessage `json:"options"`
	Correct     json.RawMessage `json:"correct" validate:"required"`
	Accept      json.RawMessage `json:"accept"`
	Explanation string          `json:"explanation"`
	Difficulty  int16           `json:"difficulty"`
}

func (h *adminHandler) createQuestion(w http.ResponseWriter, r *http.Request) {
	var req createQuestionReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	cid, _ := uuid.Parse(req.CategoryID)
	var expl *string
	if req.Explanation != "" {
		expl = &req.Explanation
	}
	diff := req.Difficulty
	if diff == 0 {
		diff = 1
	}
	q, err := h.q.CreateQuestion(r.Context(), store.CreateQuestionParams{
		CategoryID: cid, Type: req.Type, Prompt: req.Prompt,
		Options: req.Options, Correct: req.Correct, Accept: req.Accept,
		Explanation: expl, Difficulty: diff,
	})
	if err != nil {
		h.logger.Error("admin createQuestion", "err", err)
		writeErr(w, http.StatusBadRequest, "savol yaratib bo'lmadi (kategoriya mavjudmi?)")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": q.ID.String()})
}

// listQuestions — kategoriya savollari (admin uchun, to'g'ri javob bilan).
func (h *adminHandler) listQuestions(w http.ResponseWriter, r *http.Request) {
	cid, err := uuid.Parse(r.URL.Query().Get("category"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "category id yaroqsiz")
		return
	}
	rows, err := h.q.ListQuestionsByCategory(r.Context(), cid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, q := range rows {
		out = append(out, map[string]any{
			"id": q.ID.String(), "type": q.Type, "prompt": q.Prompt,
			"options": json.RawMessage(q.Options), "correct": json.RawMessage(q.Correct),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *adminHandler) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id yaroqsiz")
		return
	}
	if err := h.q.DeleteQuestion(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, "o'chirib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
