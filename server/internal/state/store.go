// Package state — jonli (ephemeral) o'yin holati. Hozir in-memory.
//
// Doimiy ma'lumotdan (Postgres) ATAYLAB ajratilgan (PLAN.md §1, §2). Scaling
// bosqichida `Store` interfeysi Redis implementatsiyasi bilan almashtiriladi —
// engine kodi o'zgarmaydi.
package state

import "sync"

type Status string

const (
	Lobby    Status = "lobby"
	Running  Status = "running"
	Finished Status = "finished"
)

// Option — variant. `Correct` faqat serverda; client'ga hech qachon yuborilmaydi.
type Option struct {
	ID      string
	Text    string
	Correct bool
}

// Question — savol bankidan kelgan shakl (provider qaytaradi).
type Question struct {
	ID          string
	Type        string
	Prompt      string
	Explanation string
	Options     []Option
}

// LiveQuestion — o'yin davomidagi savol holati.
type LiveQuestion struct {
	Question
	AskedAt  int64           // epoch ms
	Deadline int64           // epoch ms
	Answered map[string]bool // userID -> javob berdi
}

// CorrectID — to'g'ri variant id'si (mcq uchun).
func (q *LiveQuestion) CorrectID() string {
	for _, o := range q.Options {
		if o.Correct {
			return o.ID
		}
	}
	return ""
}

type Player struct {
	UserID     string
	Name       string
	Score      float64
	CorrectCnt int
	Connected  bool
	IsBot      bool
	JoinedAt   int64
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
	Questions  []*LiveQuestion
	CurrentIdx int
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
