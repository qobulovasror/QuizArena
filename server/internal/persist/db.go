// Package persist — o'yin natijasini Postgres'ga yozadi (game.Persister implementatsiyasi).
package persist

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/game"
	"github.com/azizbek12234/quizarena/server/internal/rating"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

type DB struct {
	q *store.Queries
}

func NewDB(q *store.Queries) *DB { return &DB{q: q} }

// SaveGame — game_sessions + game_results yozadi (answers_log keyin, savollar DB'da bo'lgach).
func (d *DB) SaveGame(ctx context.Context, rec game.GameRecord) error {
	subj, err := d.q.GetSubjectBySlug(ctx, rec.SubjectSlug)
	if err != nil {
		return fmt.Errorf("soha topilmadi (%q): %w", rec.SubjectSlug, err)
	}
	hostID, err := uuid.Parse(rec.HostUserID)
	if err != nil {
		return fmt.Errorf("host id yaroqsiz: %w", err)
	}

	started, finished := rec.StartedAt, rec.FinishedAt
	sess, err := d.q.InsertFinishedSession(ctx, store.InsertFinishedSessionParams{
		Code:          rec.Code,
		HostUserID:    hostID,
		SubjectID:     subj.ID,
		CategoryID:    nil,
		Mode:          rec.Mode,
		Opponent:      rec.Opponent,
		QuestionCount: int32(rec.QuestionCount),
		TimePerQ:      int32(rec.TimePerQ),
		StartedAt:     &started,
		FinishedAt:    &finished,
	})
	if err != nil {
		return fmt.Errorf("sessiya yozish: %w", err)
	}

	for _, r := range rec.Results {
		uid, err := uuid.Parse(r.UserID)
		if err != nil {
			continue // anonim/yaroqsiz id — o'tkazib yuborish
		}
		rank := int32(r.Rank)
		if _, err := d.q.UpsertResult(ctx, store.UpsertResultParams{
			SessionID:  sess.ID,
			UserID:     uid,
			Score:      r.Score,
			CorrectCnt: int32(r.CorrectCnt),
			Rank:       &rank,
		}); err != nil {
			return fmt.Errorf("natija yozish (%s): %w", r.UserID, err)
		}
	}

	// answers_log — analitika / anti-cheat audit. question_id FK questions(id)
	// bo'lgani uchun faqat DB-bankdagi savollar (general provider) yoziladi;
	// generativ savollar (UUID emas) o'tkazib yuboriladi.
	for _, a := range rec.Answers {
		uid, err := uuid.Parse(a.UserID)
		if err != nil {
			continue
		}
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil {
			continue // generativ savol — DB'da yo'q
		}
		if err := d.q.InsertAnswerLog(ctx, store.InsertAnswerLogParams{
			SessionID:  sess.ID,
			UserID:     uid,
			QuestionID: qid,
			Given:      a.Given,
			IsCorrect:  a.IsCorrect,
			TimeMs:     int32(a.TimeMs),
		}); err != nil {
			return fmt.Errorf("javob logini yozish: %w", err)
		}
	}

	// 🏆 1v1 duel — ELO reytingini yangilash (aynan 2 persistent natija bo'lsa).
	if rec.Ranked && len(rec.Results) == 2 {
		if err := d.applyElo(ctx, subj.ID, rec.Results); err != nil {
			return fmt.Errorf("reyting yangilash: %w", err)
		}
	}
	return nil
}

// applyElo — duel natijasiga ko'ra ikki o'yinchining subject reytingini yangilaydi.
func (d *DB) applyElo(ctx context.Context, subjectID uuid.UUID, results []game.ResultRecord) error {
	a, b := results[0], results[1]
	aid, err1 := uuid.Parse(a.UserID)
	bid, err2 := uuid.Parse(b.UserID)
	if err1 != nil || err2 != nil {
		return nil // anonim — ELO yo'q
	}
	ra := d.ratingOf(ctx, aid, subjectID)
	rb := d.ratingOf(ctx, bid, subjectID)

	var sa, sb float64
	switch {
	case a.Score > b.Score:
		sa, sb = 1, 0
	case a.Score < b.Score:
		sa, sb = 0, 1
	default:
		sa, sb = 0.5, 0.5
	}

	if err := d.q.UpsertRating(ctx, store.UpsertRatingParams{UserID: aid, SubjectID: subjectID, Rating: int32(rating.Next(ra, rb, sa))}); err != nil {
		return err
	}
	return d.q.UpsertRating(ctx, store.UpsertRatingParams{UserID: bid, SubjectID: subjectID, Rating: int32(rating.Next(rb, ra, sb))})
}

// ratingOf — o'yinchining joriy reytingi (yozuv yo'q bo'lsa boshlang'ich).
func (d *DB) ratingOf(ctx context.Context, userID, subjectID uuid.UUID) int {
	r, err := d.q.GetRating(ctx, store.GetRatingParams{UserID: userID, SubjectID: subjectID})
	if err != nil {
		return rating.Default
	}
	return int(r.Rating)
}
