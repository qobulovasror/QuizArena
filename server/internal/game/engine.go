// Package game — server-authoritative o'yin engine'i (classic mode, Bosqich 1).
package game

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/state"
	"github.com/azizbek12234/quizarena/server/internal/ws"
)

// Provider — savol manbai (providers.Sample, keyin EnglishVerbProvider, ...).
type Provider interface {
	Questions(count int) ([]state.Question, error)
}

// Engine — o'yin oqimini boshqaradi. Transport (ws.Hub) va jonli holat (state.Store)
// ustida ishlaydi; doimiy ma'lumot (DB) keyin (auth bilan) ulanadi.
type Engine struct {
	hub       *ws.Hub
	store     state.Store
	registry  *Registry
	persister Persister // nil bo'lsa saqlamaydi
	logger    *slog.Logger

	Countdown int           // boshlanish sanog'i (soniya); testda 0
	RevealGap time.Duration // reveal'dan keyingi pauza
}

func NewEngine(hub *ws.Hub, store state.Store, registry *Registry, persister Persister, logger *slog.Logger) *Engine {
	return &Engine{
		hub:       hub,
		store:     store,
		registry:  registry,
		persister: persister,
		logger:    logger,
		Countdown: 5,
		RevealGap: 2 * time.Second,
	}
}

// ---- Xona hayot sikli ----

func (e *Engine) CreateRoom(c *ws.Client, d ws.RoomCreateData) {
	if d.QuestionCount < 1 || d.QuestionCount > 100 {
		c.SendError(ws.ErrBadRequest, "questionCount 1..100 bo'lishi kerak")
		return
	}
	if d.TimePerQ < 1 || d.TimePerQ > 300 {
		c.SendError(ws.ErrBadRequest, "timePerQ 1..300 (soniya) bo'lishi kerak")
		return
	}
	if d.DisplayName == "" {
		c.SendError(ws.ErrBadRequest, "displayName kerak")
		return
	}
	mode := d.Mode
	if mode == "" {
		mode = "classic"
	}

	sessionID := uuid.NewString()
	userID := playerID(c)
	room := &state.Room{
		SessionID: sessionID,
		Code:      genCode(),
		HostID:    userID,
		Status:    state.Lobby,
		Config: state.Config{
			SubjectID: d.SubjectID, CategoryID: d.CategoryID, Mode: mode,
			Opponent: orDefault(d.Opponent, "human"), QuestionCount: d.QuestionCount, TimePerQ: d.TimePerQ,
		},
		Players: map[string]*state.Player{
			userID: {UserID: userID, Name: d.DisplayName, Connected: true, JoinedAt: nowMs(), Persistent: c.AuthUserID() != ""},
		},
	}
	e.store.Create(room)

	c.SetIdentity(userID, d.DisplayName)
	e.hub.Join(c, sessionID)
	c.Send(ws.SRoomJoined, ws.RoomJoinedData{SessionID: sessionID, UserID: userID, ResumeToken: userID})
	e.broadcastState(room)
}

func (e *Engine) JoinRoom(c *ws.Client, d ws.RoomJoinData) {
	if d.Code == "" || d.DisplayName == "" {
		c.SendError(ws.ErrBadRequest, "code va displayName kerak")
		return
	}
	room, ok := e.store.ByCode(d.Code)
	if !ok {
		c.SendError(ws.ErrRoomNotFound, "xona topilmadi")
		return
	}
	room.Mu.Lock()
	if room.Status != state.Lobby {
		room.Mu.Unlock()
		c.SendError(ws.ErrGameAlreadyStarted, "o'yin allaqachon boshlangan")
		return
	}
	userID := playerID(c)
	room.Players[userID] = &state.Player{UserID: userID, Name: d.DisplayName, Connected: true, JoinedAt: nowMs(), Persistent: c.AuthUserID() != ""}
	sessionID := room.SessionID
	room.Mu.Unlock()

	c.SetIdentity(userID, d.DisplayName)
	e.hub.Join(c, sessionID)
	c.Send(ws.SRoomJoined, ws.RoomJoinedData{SessionID: sessionID, UserID: userID, ResumeToken: userID})
	e.broadcastState(room)
}

func (e *Engine) Resume(c *ws.Client, d ws.RoomResumeData) {
	room, ok := e.store.Get(d.SessionID)
	if !ok {
		c.SendError(ws.ErrRoomNotFound, "sessiya topilmadi")
		return
	}
	room.Mu.Lock()
	p, ok := room.Players[d.ResumeToken]
	if !ok {
		room.Mu.Unlock()
		c.SendError(ws.ErrUnauthenticated, "resumeToken yaroqsiz")
		return
	}
	p.Connected = true
	name := p.Name
	running := room.Status == state.Running
	var cur *state.LiveQuestion
	if running && room.CurrentIdx < len(room.Questions) {
		cur = room.Questions[room.CurrentIdx]
	}
	idx, total := room.CurrentIdx, len(room.Questions)
	room.Mu.Unlock()

	c.SetIdentity(d.ResumeToken, name)
	e.hub.Join(c, room.SessionID)
	c.Send(ws.SRoomJoined, ws.RoomJoinedData{SessionID: room.SessionID, UserID: d.ResumeToken, ResumeToken: d.ResumeToken})
	e.broadcastState(room)
	if cur != nil {
		c.Send(ws.SQuestionShow, e.showPayload(cur, idx, total)) // qolgan deadline bilan
	}
}

func (e *Engine) StartGame(c *ws.Client) {
	room, ok := e.store.Get(c.Room())
	if !ok {
		c.SendError(ws.ErrRoomNotFound, "xona topilmadi")
		return
	}
	room.Mu.Lock()
	if c.UserID() != room.HostID {
		room.Mu.Unlock()
		c.SendError(ws.ErrNotHost, "faqat host boshlay oladi")
		return
	}
	if room.Status != state.Lobby {
		room.Mu.Unlock()
		c.SendError(ws.ErrGameAlreadyStarted, "o'yin allaqachon boshlangan")
		return
	}
	qs, err := e.registry.For(room.Config.SubjectID).Questions(room.Config.QuestionCount)
	if err != nil || len(qs) == 0 {
		room.Mu.Unlock()
		c.SendError(ws.ErrInternal, "savollarni olishda xato")
		return
	}
	room.Questions = buildLive(qs)
	room.Status = state.Running
	room.CurrentIdx = 0
	room.StartedAt = time.Now()
	room.Mu.Unlock()

	go e.run(room)
}

func (e *Engine) SubmitAnswer(c *ws.Client, d ws.AnswerSubmitData) {
	room, ok := e.store.Get(c.Room())
	if !ok {
		c.SendError(ws.ErrRoomNotFound, "xona topilmadi")
		return
	}
	userID := c.UserID()

	room.Mu.Lock()
	if room.Status != state.Running {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "o'yin faol emas")
		return
	}
	if d.QuestionIndex != room.CurrentIdx {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "savol indeksi mos emas")
		return
	}
	q := room.Questions[room.CurrentIdx]
	now := nowMs()
	if now > q.Deadline {
		room.Mu.Unlock()
		c.SendError(ws.ErrDeadlinePassed, "vaqt tugadi")
		return
	}
	if q.Answered[userID] {
		room.Mu.Unlock()
		c.SendError(ws.ErrAlreadyAnswered, "javob berilgan")
		return
	}
	q.Answered[userID] = true

	correct := isCorrect(q, d.Choice)
	if p := room.Players[userID]; p != nil {
		if correct {
			p.Score += scoreFor(q, now)
			p.CorrectCnt++
		}
	}
	room.Mu.Unlock()

	// To'g'rilikni OSHKOR QILMAYDI — faqat qabul qilingani.
	c.Send(ws.SAnswerAck, ws.AnswerAckData{Index: d.QuestionIndex, Accepted: true})
	// TODO(auth bilan): answers_log DB'ga yoziladi.
}

// HandleDisconnect — ulanish uzilganda o'yinchini belgilaydi/o'chiradi.
func (e *Engine) HandleDisconnect(c *ws.Client) {
	room, ok := e.store.Get(c.Room())
	if !ok {
		return
	}
	room.Mu.Lock()
	empty := false
	if room.Status == state.Lobby {
		delete(room.Players, c.UserID())
		empty = len(room.Players) == 0
	} else if p := room.Players[c.UserID()]; p != nil {
		p.Connected = false
	}
	room.Mu.Unlock()

	if empty {
		e.store.Delete(room.SessionID)
		return
	}
	e.broadcastState(room)
}

func (e *Engine) Leave(c *ws.Client) {
	e.HandleDisconnect(c)
	e.hub.Leave(c)
}

// ---- O'yin yurituvchi (har xona uchun bitta goroutine) ----

func (e *Engine) run(room *state.Room) {
	sessionID := room.SessionID

	for s := e.Countdown; s > 0; s-- {
		e.hub.BroadcastMsg(sessionID, ws.SGameCountdown, ws.CountdownData{SecondsLeft: s})
		time.Sleep(time.Second)
	}

	room.Mu.RLock()
	total := len(room.Questions)
	timePerQ := room.Config.TimePerQ
	room.Mu.RUnlock()

	for idx := 0; idx < total; idx++ {
		room.Mu.Lock()
		room.CurrentIdx = idx
		q := room.Questions[idx]
		q.AskedAt = nowMs()
		q.Deadline = q.AskedAt + int64(timePerQ)*1000
		deadline := q.Deadline
		room.Mu.Unlock()

		e.hub.BroadcastMsg(sessionID, ws.SQuestionShow, e.showPayload(q, idx, total))

		// Deadline'gacha kutish. TODO: hamma javob bersa erta o'tish (early-advance).
		time.Sleep(time.Until(time.UnixMilli(deadline)))

		e.hub.BroadcastMsg(sessionID, ws.SQuestionReveal, ws.QuestionRevealData{
			Index:       idx,
			Correct:     mustJSON(map[string]string{"optionId": q.CorrectID()}),
			Explanation: q.Explanation,
			Leaderboard: e.leaderboard(room),
		})
		time.Sleep(e.RevealGap)
	}

	room.Mu.Lock()
	room.Status = state.Finished
	room.Mu.Unlock()

	e.hub.BroadcastMsg(sessionID, ws.SGameOver, ws.GameOverData{FinalLeaderboard: e.leaderboard(room)})

	e.persist(room)

	// Tugagandan keyin biroz turadi (kech reconnect natijani ko'rsin), so'ng tozalanadi.
	time.AfterFunc(60*time.Second, func() { e.store.Delete(sessionID) })
}

// persist — o'yin natijasini doimiy saqlaydi (faqat tokenli/haqiqiy o'yinchilar).
func (e *Engine) persist(room *state.Room) {
	if e.persister == nil {
		return
	}
	room.Mu.RLock()
	host := room.Players[room.HostID]
	hostPersistent := host != nil && host.Persistent
	rec := GameRecord{
		Code: room.Code, HostUserID: room.HostID, SubjectSlug: room.Config.SubjectID,
		Mode: room.Config.Mode, Opponent: room.Config.Opponent,
		QuestionCount: room.Config.QuestionCount, TimePerQ: room.Config.TimePerQ,
		StartedAt: room.StartedAt, FinishedAt: time.Now(),
	}
	persistentOf := make(map[string]bool, len(room.Players))
	for id, p := range room.Players {
		persistentOf[id] = p.Persistent
	}
	room.Mu.RUnlock()

	if !hostPersistent {
		e.logger.Info("o'yin saqlanmadi (anonim host)", "code", rec.Code)
		return
	}
	for _, entry := range e.leaderboard(room) {
		if persistentOf[entry.UserID] {
			rec.Results = append(rec.Results, ResultRecord{
				UserID: entry.UserID, Score: entry.Score, CorrectCnt: entry.CorrectCnt, Rank: entry.Rank,
			})
		}
	}
	if err := e.persister.SaveGame(context.Background(), rec); err != nil {
		e.logger.Error("o'yinni saqlash", "err", err, "code", rec.Code)
	}
}

// ---- Yordamchilar ----

func (e *Engine) showPayload(q *state.LiveQuestion, idx, total int) ws.QuestionShowData {
	opts := make([]ws.Option, len(q.Options))
	for i, o := range q.Options {
		opts[i] = ws.Option{ID: o.ID, Text: o.Text} // Correct YO'Q
	}
	return ws.QuestionShowData{
		Index: idx, Total: total, Type: q.Type, Prompt: q.Prompt,
		Options: opts, DeadlineTs: q.Deadline,
	}
}

func (e *Engine) broadcastState(room *state.Room) {
	room.Mu.RLock()
	st := ws.RoomStateData{
		SessionID: room.SessionID,
		Code:      room.Code,
		Host:      room.HostID,
		Status:    string(room.Status),
		Config: ws.RoomConfig{
			SubjectID: room.Config.SubjectID, Mode: room.Config.Mode,
			QuestionCount: room.Config.QuestionCount, TimePerQ: room.Config.TimePerQ,
		},
		Players: make([]ws.Player, 0, len(room.Players)),
	}
	for _, p := range room.Players {
		st.Players = append(st.Players, ws.Player{
			UserID: p.UserID, Name: p.Name, Score: p.Score, Connected: p.Connected, IsBot: p.IsBot,
		})
	}
	sessionID := room.SessionID
	room.Mu.RUnlock()
	e.hub.BroadcastMsg(sessionID, ws.SRoomState, st)
}

func (e *Engine) leaderboard(room *state.Room) []ws.LeaderboardEntry {
	room.Mu.RLock()
	entries := make([]ws.LeaderboardEntry, 0, len(room.Players))
	for _, p := range room.Players {
		entries = append(entries, ws.LeaderboardEntry{
			UserID: p.UserID, Name: p.Name, Score: p.Score, CorrectCnt: p.CorrectCnt,
		})
	}
	room.Mu.RUnlock()

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Score > entries[j].Score })
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries
}

// scoreFor — ball: 100 baza + tezlik bonusi (0..100). Faqat to'g'ri javobga.
func scoreFor(q *state.LiveQuestion, now int64) float64 {
	const base = 100.0
	total := float64(q.Deadline - q.AskedAt)
	if total <= 0 {
		return base
	}
	remaining := float64(q.Deadline - now)
	if remaining < 0 {
		remaining = 0
	}
	return base + base*(remaining/total)
}

// isCorrect — mcq uchun tanlangan optionId to'g'rimi (server-authoritative).
func isCorrect(q *state.LiveQuestion, choice json.RawMessage) bool {
	var ch struct {
		OptionID string `json:"optionId"`
	}
	if err := json.Unmarshal(choice, &ch); err != nil {
		return false
	}
	for _, o := range q.Options {
		if o.ID == ch.OptionID {
			return o.Correct
		}
	}
	return false
}

func buildLive(qs []state.Question) []*state.LiveQuestion {
	live := make([]*state.LiveQuestion, len(qs))
	for i, q := range qs {
		opts := append([]state.Option(nil), q.Options...)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		q.Options = opts
		live[i] = &state.LiveQuestion{Question: q, Answered: make(map[string]bool)}
	}
	return live
}

func genCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func nowMs() int64 { return time.Now().UnixMilli() }

// playerID — autentifikatsiyalangan userID bo'lsa shuni, aks holda yangi guest id.
func playerID(c *ws.Client) string {
	if id := c.AuthUserID(); id != "" {
		return id
	}
	return uuid.NewString()
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
