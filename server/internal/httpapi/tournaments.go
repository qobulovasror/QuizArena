package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/azizbek12234/quizarena/server/internal/game/qtype"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

type tournamentHandler struct {
	q        *store.Queries
	validate *validator.Validate
	logger   *slog.Logger
}

// status — turnir holatini vaqt bo'yicha hisoblaydi (DBda saqlanmaydi).
// now < startsAt → "upcoming"; now > endsAt → "finished"; aks holda "active".
func status(now, startsAt, endsAt time.Time) string {
	switch {
	case now.Before(startsAt):
		return "upcoming"
	case now.After(endsAt):
		return "finished"
	default:
		return "active"
	}
}

type tournamentDTO struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Subject       string `json:"subject"`
	SubjectName   string `json:"subjectName"`
	QuestionCount int32  `json:"questionCount"`
	StartsAt      string `json:"startsAt"`
	EndsAt        string `json:"endsAt"`
	Status        string `json:"status"`
}

// list — barcha turnirlar (status bilan).
func (h *tournamentHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListTournaments(r.Context())
	if err != nil {
		h.logger.Error("tournament list", "err", err)
		writeErr(w, http.StatusInternalServerError, "turnirlarni olishda xato")
		return
	}
	now := time.Now()
	out := make([]tournamentDTO, 0, len(rows))
	for _, t := range rows {
		out = append(out, tournamentDTO{
			ID: t.ID.String(), Title: t.Title, Subject: t.SubjectSlug, SubjectName: t.SubjectName,
			QuestionCount: t.QuestionCount,
			StartsAt:      t.StartsAt.Format(time.RFC3339), EndsAt: t.EndsAt.Format(time.RFC3339),
			Status: status(now, t.StartsAt, t.EndsAt),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// play — faol turnir savollari (javobsiz, opaque — assess.go uslubida).
func (h *tournamentHandler) play(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id yaroqsiz")
		return
	}
	t, err := h.q.GetTournament(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "turnir topilmadi")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	if status(time.Now(), t.StartsAt, t.EndsAt) != "active" {
		writeErr(w, http.StatusBadRequest, "turnir faol emas")
		return
	}
	rows, err := h.q.RandomQuestionsBySubject(r.Context(), store.RandomQuestionsBySubjectParams{
		SubjectID: t.SubjectID, Limit: t.QuestionCount,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "savollarni olishda xato")
		return
	}
	out := make([]assessQ, 0, len(rows))
	for _, q := range rows {
		// Variantlarni serverda aralashtiramiz — tartibga-sezgir turlarda (ordering)
		// DB tartibi to'g'ri javobni oshkor qilmasligi uchun.
		out = append(out, assessQ{QuestionID: q.ID.String(), Type: q.Type, Prompt: q.Prompt, Options: shuffleOptions(q.Options)})
	}
	writeJSON(w, http.StatusOK, out)
}

// shuffleOptions — variant massivini (opaque {id,text}) aralashtiradi.
func shuffleOptions(raw []byte) []byte {
	if len(raw) == 0 {
		return raw
	}
	var opts []json.RawMessage
	if json.Unmarshal(raw, &opts) != nil || len(opts) < 2 {
		return raw
	}
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	out, err := json.Marshal(opts)
	if err != nil {
		return raw
	}
	return out
}

// submit — javoblarni server-side baholaydi va eng yaxshi ballni saqlaydi.
func (h *tournamentHandler) submit(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(userIDFrom(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user id yaroqsiz")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id yaroqsiz")
		return
	}
	t, err := h.q.GetTournament(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "turnir topilmadi")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	if status(time.Now(), t.StartsAt, t.EndsAt) != "active" {
		writeErr(w, http.StatusBadRequest, "turnir faol emas")
		return
	}

	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) == 0 {
		writeErr(w, http.StatusBadRequest, "javoblar yo'q")
		return
	}

	// Anti-cheat: har savol bir marta, FAQAT turnir sohasidagi savollar, eng ko'pi
	// bilan questionCount tasi hisoblanadi (boshqa soha/takror ID bilan ball shishirib bo'lmaydi).
	seen := map[uuid.UUID]bool{}
	correct, counted := 0, 0
	for _, a := range req.Answers {
		if counted >= int(t.QuestionCount) {
			break
		}
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil || seen[qid] {
			continue
		}
		q, err := h.q.GetQuestionInSubject(r.Context(), store.GetQuestionInSubjectParams{ID: qid, SubjectID: t.SubjectID})
		if err != nil {
			continue // boshqa soha yoki mavjud emas — e'tiborsiz
		}
		seen[qid] = true
		counted++
		if qtype.For(q.Type).Validate(a.Choice, q.Correct) {
			correct++
		}
	}

	if err := h.q.UpsertEntry(r.Context(), store.UpsertEntryParams{
		TournamentID: id, UserID: uid, Score: int32(correct),
	}); err != nil {
		h.logger.Error("tournament upsert entry", "err", err)
		writeErr(w, http.StatusInternalServerError, "natijani saqlashda xato")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"correct": correct, "total": len(req.Answers)})
}

type entryDTO struct {
	Username string `json:"username"`
	Score    int32  `json:"score"`
}

// leaderboard — turnir reytingi (ball bo'yicha tartiblangan).
func (h *tournamentHandler) leaderboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id yaroqsiz")
		return
	}
	rows, err := h.q.ListEntries(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "reytingni olishda xato")
		return
	}
	out := make([]entryDTO, 0, len(rows))
	for _, e := range rows {
		name := ""
		if e.Username != nil {
			name = *e.Username
		}
		out = append(out, entryDTO{Username: name, Score: e.Score})
	}
	writeJSON(w, http.StatusOK, out)
}

type createTournamentReq struct {
	Title         string `json:"title" validate:"required"`
	SubjectSlug   string `json:"subjectSlug" validate:"required"`
	QuestionCount int32  `json:"questionCount"`
	StartsAt      string `json:"startsAt" validate:"required"`
	EndsAt        string `json:"endsAt" validate:"required"`
}

// create — admin uchun turnir yaratish.
func (h *tournamentHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createTournamentReq
	if !decodeValidate(w, r, h.validate, &req) {
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "startsAt RFC3339 bo'lishi kerak")
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "endsAt RFC3339 bo'lishi kerak")
		return
	}
	if !endsAt.After(startsAt) {
		writeErr(w, http.StatusBadRequest, "endsAt startsAt'dan keyin bo'lishi kerak")
		return
	}
	subj, err := h.q.GetSubjectBySlug(r.Context(), req.SubjectSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "soha topilmadi")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "xato")
		return
	}
	qc := req.QuestionCount
	if qc <= 0 {
		qc = 10
	}
	t, err := h.q.CreateTournament(r.Context(), store.CreateTournamentParams{
		Title: req.Title, SubjectID: subj.ID, QuestionCount: qc, StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		h.logger.Error("tournament create", "err", err)
		writeErr(w, http.StatusBadRequest, "turnir yaratib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": t.ID.String()})
}
