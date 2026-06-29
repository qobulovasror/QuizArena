package game

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/state"
	"github.com/azizbek12234/quizarena/server/internal/ws"
)

// duelHuman — duel ishtirokchisi (inson). uid startDuel ichida to'ldiriladi.
type duelHuman struct {
	client *ws.Client
	name   string
	uid    string
}

// Matchmaker — 🏆 1v1 navbat (subject bo'yicha bitta kutuvchi slot). Raqib topilmasa
// botDelay'dan keyin bot bilan duel boshlanadi (PLAN §1.7 — navbat hech qachon bo'sh qolmaydi).
type Matchmaker struct {
	mu       sync.Mutex
	waiting  map[string]*waiter // subjectID(slug) -> kutuvchi
	engine   *Engine
	botDelay time.Duration
}

type waiter struct {
	client *ws.Client
	name   string
	timer  *time.Timer
}

func NewMatchmaker(e *Engine) *Matchmaker {
	return &Matchmaker{waiting: map[string]*waiter{}, engine: e, botDelay: 10 * time.Second}
}

// Queue — o'yinchini navbatga qo'shadi; kutuvchi bo'lsa darhol juftlaydi.
func (m *Matchmaker) Queue(c *ws.Client, subjectID, name string) {
	if subjectID == "" {
		c.SendError(ws.ErrBadRequest, "subjectId kerak")
		return
	}
	if name == "" {
		name = "O'yinchi"
	}
	m.mu.Lock()
	if w, ok := m.waiting[subjectID]; ok {
		if w.client == c { // shu mijoz allaqachon kutyapti
			m.mu.Unlock()
			return
		}
		w.timer.Stop()
		delete(m.waiting, subjectID)
		m.mu.Unlock()
		m.engine.startDuel(subjectID, []duelHuman{{client: w.client, name: w.name}, {client: c, name: name}}, false)
		return
	}
	t := time.AfterFunc(m.botDelay, func() { m.fallbackBot(subjectID, c) })
	m.waiting[subjectID] = &waiter{client: c, name: name, timer: t}
	m.mu.Unlock()
	c.Send(ws.SMatchQueued, ws.MatchQueuedData{SubjectID: subjectID})
}

// fallbackBot — kutish vaqti tugadi: bitta inson + bot duel.
func (m *Matchmaker) fallbackBot(subjectID string, c *ws.Client) {
	m.mu.Lock()
	w, ok := m.waiting[subjectID]
	if !ok || w.client != c {
		m.mu.Unlock()
		return
	}
	name := w.name
	delete(m.waiting, subjectID)
	m.mu.Unlock()
	m.engine.startDuel(subjectID, []duelHuman{{client: c, name: name}}, true)
}

// Cancel — mijozni barcha navbatlardan chiqaradi (match:cancel / uzilish).
func (m *Matchmaker) Cancel(c *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for sid, w := range m.waiting {
		if w.client == c {
			w.timer.Stop()
			delete(m.waiting, sid)
		}
	}
}

// startDuel — 1v1 duel xonasi (classic, 5 savol, 15s) yaratadi va darhol boshlaydi.
// withBot=true → bitta inson + bot; aks holda ikki inson (Ranked → ELO).
func (e *Engine) startDuel(subjectID string, humans []duelHuman, withBot bool) {
	qs, err := e.registry.For(subjectID).Questions(5)
	if err != nil || len(qs) == 0 {
		for _, h := range humans {
			h.client.SendError(ws.ErrInternal, "duel savollari topilmadi")
		}
		return
	}
	opponent := "human"
	if withBot {
		opponent = "bot"
	}
	sessionID := uuid.NewString()
	room := &state.Room{
		SessionID: sessionID, Code: genCode(), Status: state.Lobby,
		Config: state.Config{
			SubjectID: subjectID, Mode: "classic", Opponent: opponent,
			QuestionCount: 5, TimePerQ: 15,
		},
		Players: map[string]*state.Player{},
		Ranked:  !withBot, // faqat inson-inson duel reytingli
	}
	for i := range humans {
		uid := playerID(humans[i].client)
		humans[i].uid = uid
		room.Players[uid] = &state.Player{
			UserID: uid, Name: humans[i].name, Connected: true, JoinedAt: nowMs(),
			Persistent: humans[i].client.AuthUserID() != "",
		}
		if i == 0 {
			room.HostID = uid
		}
	}
	if withBot {
		addBot(room)
	}
	room.Questions = buildLive(qs)
	room.Status = state.Running
	room.StartedAt = time.Now()
	e.store.Create(room)

	for _, h := range humans {
		h.client.SetIdentity(h.uid, h.name)
		e.hub.Join(h.client, sessionID)
		h.client.Send(ws.SMatchFound, ws.MatchFoundData{SessionID: sessionID, VsBot: withBot})
		h.client.Send(ws.SRoomJoined, ws.RoomJoinedData{SessionID: sessionID, UserID: h.uid, ResumeToken: h.uid})
	}
	e.broadcastState(room)
	go e.run(room)
}
