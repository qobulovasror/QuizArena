package game

import (
	"context"
	"encoding/json"
	"time"
)

// Persister — o'yin tugagach doimiy ma'lumotni (tarix) saqlaydi.
// Engine'dan ajratilgan (interfeys) — DB'siz testlanadi; nil bo'lsa saqlamaydi.
type Persister interface {
	SaveGame(ctx context.Context, rec GameRecord) error
}

type GameRecord struct {
	Code          string
	HostUserID    string
	SubjectSlug   string
	Mode          string
	Opponent      string
	QuestionCount int
	TimePerQ      int
	StartedAt     time.Time
	FinishedAt    time.Time
	Ranked        bool // 🏆 1v1 duel → ELO yangilanadi (aynan 2 persistent natija bo'lsa)
	Results       []ResultRecord
	Answers       []AnswerRecord
}

type ResultRecord struct {
	UserID     string
	Score      float64
	CorrectCnt int
	Rank       int
}

// AnswerRecord — answers_log uchun bitta javob (analitika / anti-cheat audit).
type AnswerRecord struct {
	UserID     string
	QuestionID string
	Given      json.RawMessage
	IsCorrect  bool
	TimeMs     int
}
