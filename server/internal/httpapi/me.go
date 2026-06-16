package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

type meHandler struct {
	q      *store.Queries
	logger *slog.Logger
}

type historyItem struct {
	Code       string     `json:"code"`
	Mode       string     `json:"mode"`
	Subject    string     `json:"subject"`
	Score      float64    `json:"score"`
	CorrectCnt int32      `json:"correctCnt"`
	Rank       *int32     `json:"rank"`
	FinishedAt *time.Time `json:"finishedAt"`
}

// history — joriy foydalanuvchining o'yin tarixi.
func (h *meHandler) history(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
	rows, err := h.q.ListHistoryByUser(r.Context(), id)
	if err != nil {
		h.logger.Error("history", "err", err)
		writeErr(w, http.StatusInternalServerError, "tarixni olishda xato")
		return
	}
	items := make([]historyItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyItem{
			Code: row.Code, Mode: row.Mode, Subject: row.SubjectSlug,
			Score: row.Score, CorrectCnt: row.CorrectCnt, Rank: row.Rank, FinishedAt: row.FinishedAt,
		})
	}
	writeJSON(w, http.StatusOK, items)
}
