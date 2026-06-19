package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/learn"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

type srsHandler struct {
	q        *store.Queries
	validate *validator.Validate
	logger   *slog.Logger
}

// Flashcard: prompt + (ko'rsatiladigan) javob + izoh. Practice yakka/o'rganish — javob OCHIQ.
type srsCardDTO struct {
	QuestionID  string `json:"questionId"`
	Type        string `json:"type"`
	Prompt      string `json:"prompt"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation,omitempty"`
}

const srsBatch = 10

// due — takror qilinishi kerak (yoki yangi) kartalar.
func (h *srsHandler) due(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
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
		h.logger.Error("srs subject", "err", err)
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}

	items := make([]srsCardDTO, 0, srsBatch)
	dueRows, err := h.q.GetDueCards(r.Context(), store.GetDueCardsParams{UserID: uid, SubjectID: subj.ID, Limit: srsBatch})
	if err != nil {
		h.logger.Error("srs due", "err", err)
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	for _, d := range dueRows {
		items = append(items, srsDTO(d.ID, d.Type, d.Prompt, d.Options, d.Correct, d.Explanation))
	}
	// Yetmasa — yangi savollar bilan to'ldiramiz.
	if len(items) < srsBatch {
		newRows, err := h.q.GetNewQuestions(r.Context(), store.GetNewQuestionsParams{
			SubjectID: subj.ID, UserID: uid, Limit: int32(srsBatch - len(items)),
		})
		if err == nil {
			for _, n := range newRows {
				items = append(items, srsDTO(n.ID, n.Type, n.Prompt, n.Options, n.Correct, n.Explanation))
			}
		}
	}
	writeJSON(w, http.StatusOK, items)
}

type reviewReq struct {
	QuestionID string `json:"questionId" validate:"required"`
	Grade      int    `json:"grade"` // 0=qaytadan, 1=yaxshi, 2=oson
}

// review — javob bahosini qabul qilib, SM-2 holatini yangilaydi.
func (h *srsHandler) review(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
	var req reviewReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	qid, err := uuid.Parse(req.QuestionID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "questionId yaroqsiz")
		return
	}

	card := learn.NewCard()
	if cur, err := h.q.GetSrsCard(r.Context(), store.GetSrsCardParams{UserID: uid, QuestionID: qid}); err == nil {
		card = learn.Card{Ease: cur.Ease, Interval: int(cur.IntervalDays), Reps: int(cur.Repetitions)}
	}
	nc, due := learn.Review(card, req.Grade, time.Now())

	if err := h.q.UpsertSrsCard(r.Context(), store.UpsertSrsCardParams{
		UserID: uid, QuestionID: qid, Ease: nc.Ease,
		IntervalDays: int32(nc.Interval), Repetitions: int32(nc.Reps), DueAt: due,
	}); err != nil {
		writeErr(w, http.StatusBadRequest, "savol topilmadi yoki yozib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nextDue": due, "intervalDays": nc.Interval})
}

func srsDTO(id uuid.UUID, typ, prompt string, options, correct []byte, expl *string) srsCardDTO {
	e := ""
	if expl != nil {
		e = *expl
	}
	return srsCardDTO{QuestionID: id.String(), Type: typ, Prompt: prompt, Answer: answerText(typ, options, correct), Explanation: e}
}

// answerText — flashcard'da ko'rsatiladigan o'qiladigan javob (tur bo'yicha).
func answerText(typ string, options, correct []byte) string {
	switch typ {
	case "true_false":
		var c struct {
			Value bool `json:"value"`
		}
		_ = json.Unmarshal(correct, &c)
		if c.Value {
			return "To'g'ri"
		}
		return "Noto'g'ri"
	case "numeric":
		var c struct {
			Value float64 `json:"value"`
		}
		_ = json.Unmarshal(correct, &c)
		return strconv.FormatFloat(c.Value, 'g', -1, 64)
	case "type_answer", "fill_blank":
		var c struct {
			Accepted []string `json:"accepted"`
		}
		_ = json.Unmarshal(correct, &c)
		return strings.Join(c.Accepted, ", ")
	default: // mcq, multi_select
		var c struct {
			OptionID string `json:"optionId"`
		}
		_ = json.Unmarshal(correct, &c)
		var opts []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		_ = json.Unmarshal(options, &opts)
		for _, o := range opts {
			if o.ID == c.OptionID {
				return o.Text
			}
		}
		return fmt.Sprintf("%s", correct)
	}
}
