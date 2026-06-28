// Package state — jonli (ephemeral) o'yin holati. Hozir in-memory.
//
// Doimiy ma'lumotdan (Postgres) ATAYLAB ajratilgan (PLAN.md §1, §2). Scaling
// bosqichida `Store` interfeysi Redis implementatsiyasi bilan almashtiriladi —
// engine kodi o'zgarmaydi.
package state

import (
	"encoding/json"
	"sync"
	"time"
)

type Status string

const (
	Lobby    Status = "lobby"
	Running  Status = "running"
	Finished Status = "finished"
)

// Option — client'ga ko'rinadigan variant (to'g'ri javob YO'Q).
type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Question — savol bankidan kelgan shakl (provider qaytaradi).
// `Correct` — tur-spetsifik to'g'ri javob (reveal shakli), FAQAT serverda.
// QuestionType strategiyasi uni Validate/Reveal uchun ishlatadi (qarang game/qtype).
type Question struct {
	ID          string
	Type        string
	Prompt      string
	Explanation string
	Options     []Option        // mcq/multi_select/ordering/match(chap)/categorize(element)
	Targets     []Option        // match(o'ng) / categorize(toifa) — aks holda bo'sh
	Correct     json.RawMessage // masalan {"optionId":"o2"} | {"value":12,"tolerance":0.5}
}

// LiveQuestion — o'yin davomidagi savol holati.
type LiveQuestion struct {
	Question
	AskedAt  int64           // epoch ms
	Deadline int64           // epoch ms
	Answered map[string]bool // userID -> javob berdi
}

type Player struct {
	UserID     string
	Name       string
	Score      float64
	CorrectCnt int
	Connected  bool
	IsBot      bool
	JoinedAt   int64
	Persistent bool   // tokenli (haqiqiy users yozuvi) → natija DB'ga yoziladi
	Eliminated bool   // survival rejimi: xato javobdan keyin o'yindan chiqdi
	Team       string // team rejimi: jamoa belgisi ("A"/"B"), StartGame'da tayinlanadi
	TaIdx      int    // time_attack: o'yinchining joriy savol indeksi (per-player oqim)
	TaDone     bool   // time_attack: barcha savollarga javob berdi
}

// AnswerEvent — bitta o'yinchi javobi, answers_log audit uchun yig'iladi.
// `Given` — xom tanlov; `QuestionID` DB-savol bo'lsa UUID (generativ savolda emas).
type AnswerEvent struct {
	UserID     string
	QuestionID string
	Given      json.RawMessage
	IsCorrect  bool
	TimeMs     int
}

type Config struct {
	SubjectID     string
	CategoryID    string
	Mode          string
	Opponent      string
	QuestionCount int
	TimePerQ      int
}

// Room — bitta o'yin xonasi. `Mu` tashqi paketdan (engine) qulflanadi.
type Room struct {
	Mu sync.RWMutex

	SessionID  string
	Code       string
	HostID     string
	Status     Status
	Config     Config
	Players    map[string]*Player
	Questions      []*LiveQuestion
	CurrentIdx     int
	StartedAt      time.Time
	GlobalDeadline int64         // time_attack: butun o'yin uchun yagona deadline (epoch ms)
	Answers        []AnswerEvent // o'yin davomida yig'iladi, tugagach answers_log'ga yoziladi
}

// Store — jonli xonalar ombori interfeysi.
type Store interface {
	Create(r *Room)
	Get(sessionID string) (*Room, bool)
	ByCode(code string) (*Room, bool)
	Delete(sessionID string)
}

// MemStore — process xotirasidagi implementatsiya.
type MemStore struct {
	mu    sync.RWMutex
	rooms map[string]*Room
	codes map[string]string // code -> sessionID
}

func NewMemStore() *MemStore {
	return &MemStore{
		rooms: make(map[string]*Room),
		codes: make(map[string]string),
	}
}

func (s *MemStore) Create(r *Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[r.SessionID] = r
	s.codes[r.Code] = r.SessionID
}

func (s *MemStore) Get(sessionID string) (*Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[sessionID]
	return r, ok
}

func (s *MemStore) ByCode(code string) (*Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.codes[code]
	if !ok {
		return nil, false
	}
	r, ok := s.rooms[id]
	return r, ok
}

func (s *MemStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rooms[sessionID]; ok {
		delete(s.codes, r.Code)
		delete(s.rooms, sessionID)
	}
}
