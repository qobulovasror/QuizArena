package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/state"
	"github.com/azizbek12234/quizarena/server/internal/store"
)

// General — statik savol bankidan (Postgres `questions`) o'qiydigan provider.
// Generativ providerlardan farqli: savollar DB'da saqlangan (admin/seed orqali).
type General struct {
	q         *store.Queries
	subjectID uuid.UUID
}

func NewGeneral(q *store.Queries, subjectID uuid.UUID) *General {
	return &General{q: q, subjectID: subjectID}
}

func (g *General) Questions(count int) ([]state.Question, error) {
	rows, err := g.q.RandomQuestionsBySubject(context.Background(), store.RandomQuestionsBySubjectParams{
		SubjectID: g.subjectID,
		Limit:     int32(count),
	})
	if err != nil {
		return nil, fmt.Errorf("savollarni olish: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("bu sohada savol yo'q")
	}
	out := make([]state.Question, 0, len(rows))
	for _, r := range rows {
		var opts []state.Option
		if len(r.Options) > 0 {
			_ = json.Unmarshal(r.Options, &opts)
		}
		expl := ""
		if r.Explanation != nil {
			expl = *r.Explanation
		}
		out = append(out, state.Question{
			ID:          r.ID.String(),
			Type:        r.Type,
			Prompt:      r.Prompt,
			Explanation: expl,
			Options:     opts,
			Correct:     json.RawMessage(r.Correct),
		})
	}
	return out, nil
}
