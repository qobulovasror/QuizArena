// Package persist — o'yin natijasini Postgres'ga yozadi (game.Persister implementatsiyasi).
package persist

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/game"
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
	return nil
}
