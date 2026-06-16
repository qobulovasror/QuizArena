package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

type subjectsHandler struct {
	q      *store.Queries
	logger *slog.Logger
}

type subjectDTO struct {
	ID   string  `json:"id"`
	Slug string  `json:"slug"`
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

type categoryDTO struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// list — barcha sohalar (ochiq, auth shart emas).
func (h *subjectsHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListSubjects(r.Context())
	if err != nil {
		h.logger.Error("subjects", "err", err)
		writeErr(w, http.StatusInternalServerError, "sohalarni olishda xato")
		return
	}
	items := make([]subjectDTO, 0, len(rows))
	for _, s := range rows {
		items = append(items, subjectDTO{ID: s.ID.String(), Slug: s.Slug, Name: s.Name, Icon: s.Icon})
	}
	writeJSON(w, http.StatusOK, items)
}

// categories — soha ichidagi kategoriyalar.
func (h *subjectsHandler) categories(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "soha id yaroqsiz")
		return
	}
	rows, err := h.q.ListCategoriesBySubject(r.Context(), id)
	if err != nil {
		h.logger.Error("categories", "err", err)
		writeErr(w, http.StatusInternalServerError, "kategoriyalarni olishda xato")
		return
	}
	items := make([]categoryDTO, 0, len(rows))
	for _, c := range rows {
		items = append(items, categoryDTO{ID: c.ID.String(), Slug: c.Slug, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, items)
}
