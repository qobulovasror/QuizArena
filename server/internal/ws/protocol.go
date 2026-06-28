// Package ws — WebSocket transport va protokol tiplari.
//
// MANBA SHARTNOMA: shared/protocol/README.md. TS tomoni: shared/protocol/messages.ts.
// Bu fayl faqat protokol *tiplari* (logika yo'q) — Hub/handlerlar keyingi bosqichda.
package ws

import "encoding/json"

// Envelope — barcha WebSocket xabarlari uchun umumiy konvert.
type Envelope struct {
	Type MsgType         `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	ID   string          `json:"id,omitempty"` // ixtiyoriy client→server korrelyatsiya
}

// MsgType — xabar turi.
type MsgType string

// Client → Server
const (
	CRoomCreate   MsgType = "room:create"
	CRoomJoin     MsgType = "room:join"
	CRoomResume   MsgType = "room:resume"
	CRoomLeave    MsgType = "room:leave"
	CGameStart    MsgType = "game:start"
	CAnswerSubmit MsgType = "answer:submit"
	CMatchQueue   MsgType = "match:queue"  // 🏆 1v1 navbatga qo'shilish
	CMatchCancel  MsgType = "match:cancel" // navbatdan chiqish
)

// Server → Client
const (
	SRoomState      MsgType = "room:state"
	SRoomJoined     MsgType = "room:joined"
	SGameCountdown  MsgType = "game:countdown"
	SQuestionShow   MsgType = "question:show"
	SAnswerAck      MsgType = "answer:ack"
	SQuestionReveal MsgType = "question:reveal"
	SPlayerScored   MsgType = "player:scored"
	SGameOver       MsgType = "game:over"
	SMatchQueued    MsgType = "match:queued" // navbatga qo'shildi (raqib kutilmoqda)
	SMatchFound     MsgType = "match:found"  // raqib topildi (duel boshlanadi)
	SError          MsgType = "error"
)

// ---- Umumiy tuzilmalar ----

type Player struct {
	UserID     string  `json:"userId"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Connected  bool    `json:"connected"`
	IsBot      bool    `json:"isBot,omitempty"`
	Eliminated bool    `json:"eliminated,omitempty"`
	Team       string  `json:"team,omitempty"` // team rejimi
}

type RoomConfig struct {
	SubjectID     string `json:"subjectId"`
	Mode          string `json:"mode"`
	QuestionCount int    `json:"questionCount"`
	TimePerQ      int    `json:"timePerQ"` // soniya
}

// Option — opaque id (server aralashtiradi), to'g'ri javob oshkor bo'lmaydi.
type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type LeaderboardEntry struct {
	UserID     string  `json:"userId"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	CorrectCnt int     `json:"correctCnt"`
	Rank       int     `json:"rank"`
	Eliminated bool    `json:"eliminated,omitempty"`
	Team       string  `json:"team,omitempty"` // team rejimi
}

// TeamStanding — jamoa yig'indisi (team rejimi; reveal/over'da yuboriladi).
type TeamStanding struct {
	Team       string  `json:"team"`
	Score      float64 `json:"score"`
	CorrectCnt int     `json:"correctCnt"`
	Rank       int     `json:"rank"`
}

// ---- Client → Server yuklar ----

type RoomCreateData struct {
	SubjectID     string `json:"subjectId"`
	CategoryID    string `json:"categoryId,omitempty"`
	Mode          string `json:"mode"`
	Opponent      string `json:"opponent,omitempty"`      // human|bot|mixed
	BotDifficulty string `json:"botDifficulty,omitempty"` // easy|medium|hard
	QuestionCount int    `json:"questionCount"`
	TimePerQ      int    `json:"timePerQ"`
	DisplayName   string `json:"displayName"` // host ham o'yinchi
}

type RoomJoinData struct {
	Code        string `json:"code"`
	DisplayName string `json:"displayName"`
}

type RoomResumeData struct {
	SessionID   string `json:"sessionId"`
	ResumeToken string `json:"resumeToken"`
}

// AnswerSubmitData — Choice xom qoladi; tur strategiyasi tekshiradi (README §6).
type AnswerSubmitData struct {
	QuestionIndex int             `json:"questionIndex"`
	Choice        json.RawMessage `json:"choice"`
}

// MatchQueueData — 1v1 navbatga qo'shilish (soha + ko'rinadigan ism).
type MatchQueueData struct {
	SubjectID   string `json:"subjectId"`
	DisplayName string `json:"displayName"`
}

// ---- Server → Client yuklar ----

type RoomStateData struct {
	SessionID string     `json:"sessionId"`
	Code      string     `json:"code"`
	Host      string     `json:"host"`
	Status    string     `json:"status"` // lobby|running|finished
	Config    RoomConfig `json:"config"`
	Players   []Player   `json:"players"`
}

type RoomJoinedData struct {
	SessionID   string `json:"sessionId"`
	UserID      string `json:"userId"`
	ResumeToken string `json:"resumeToken"`
}

type CountdownData struct {
	SecondsLeft int `json:"secondsLeft"`
}

type QuestionShowData struct {
	Index      int      `json:"index"`
	Total      int      `json:"total"`
	Type       string   `json:"type"`
	Prompt     string   `json:"prompt"`
	Options    []Option `json:"options,omitempty"`
	Targets    []Option `json:"targets,omitempty"` // match(o'ng)/categorize(toifa)
	DeadlineTs int64    `json:"deadlineTs"`        // server epoch ms
}

type AnswerAckData struct {
	Index    int  `json:"index"`
	Accepted bool `json:"accepted"` // to'g'rilikni OSHKOR QILMAYDI
}

type QuestionRevealData struct {
	Index       int                `json:"index"`
	Correct     json.RawMessage    `json:"correct"` // tur bo'yicha shakl
	Explanation string             `json:"explanation,omitempty"`
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	Teams       []TeamStanding     `json:"teams,omitempty"` // team rejimi
}

type PlayerScoredData struct {
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
}

type GameOverData struct {
	FinalLeaderboard []LeaderboardEntry `json:"finalLeaderboard"`
	Teams            []TeamStanding     `json:"teams,omitempty"` // team rejimi
}

// MatchQueuedData — navbatga qo'shilgani tasdiq (raqib kutilmoqda).
type MatchQueuedData struct {
	SubjectID string `json:"subjectId"`
}

// MatchFoundData — raqib topildi; duel xonasi (room:joined/state ketidan keladi).
type MatchFoundData struct {
	SessionID string `json:"sessionId"`
	VsBot     bool   `json:"vsBot"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Xato kodlari (README §7).
const (
	ErrBadRequest         = "BAD_REQUEST"
	ErrInvalidMessage     = "INVALID_MESSAGE"
	ErrRoomNotFound       = "ROOM_NOT_FOUND"
	ErrRoomFull           = "ROOM_FULL"
	ErrNotHost            = "NOT_HOST"
	ErrGameAlreadyStarted = "GAME_ALREADY_STARTED"
	ErrAlreadyAnswered    = "ALREADY_ANSWERED"
	ErrDeadlinePassed     = "DEADLINE_PASSED"
	ErrUnauthenticated    = "UNAUTHENTICATED"
	ErrInternal           = "INTERNAL"
)
