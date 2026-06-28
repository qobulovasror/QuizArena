// Package game — server-authoritative o'yin engine'i (classic mode, Bosqich 1).
package game

import (
	"context"
	"log/slog"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/azizbek12234/quizarena/server/internal/game/modes"
	"github.com/azizbek12234/quizarena/server/internal/game/qtype"
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

	BotCorrectProb float64 // 🏆 simulyatsion raqib to'g'ri javob ehtimoli (0..1)
}

func NewEngine(hub *ws.Hub, store state.Store, registry *Registry, persister Persister, logger *slog.Logger) *Engine {
	return &Engine{
		hub:            hub,
		store:          store,
		registry:       registry,
		persister:      persister,
		logger:         logger,
		Countdown:      5,
		RevealGap:      2 * time.Second,
		BotCorrectProb: 0.65, // o'rta qiyinlik
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
	if wantsBot(mode, room.Config.Opponent) { // 🏆 raqib bot — lobby'da ko'rinadi
		addBot(room)
	}

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
	total := len(room.Questions)
	var cur *state.LiveQuestion
	var idx int
	var deadline int64
	if running {
		if room.Config.Mode == "time_attack" {
			if !p.TaDone && p.TaIdx < total { // o'yinchining joriy savoli, yagona deadline bilan
				cur, idx, deadline = room.Questions[p.TaIdx], p.TaIdx, room.GlobalDeadline
			}
		} else if room.CurrentIdx < total {
			cur, idx, deadline = room.Questions[room.CurrentIdx], room.CurrentIdx, room.Questions[room.CurrentIdx].Deadline
		}
	}
	room.Mu.Unlock()

	c.SetIdentity(d.ResumeToken, name)
	e.hub.Join(c, room.SessionID)
	c.Send(ws.SRoomJoined, ws.RoomJoinedData{SessionID: room.SessionID, UserID: d.ResumeToken, ResumeToken: d.ResumeToken})
	e.broadcastState(room)
	if cur != nil {
		c.Send(ws.SQuestionShow, e.showPayload(cur, idx, total, deadline)) // qolgan deadline bilan
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
	if room.Config.Mode == "team" {
		assignTeams(room) // lock ushlab turilgan
	}
	room.Mu.Unlock()

	e.broadcastState(room) // status=running → clientlar Play ekraniga o'tadi
	go e.run(room)
}

func (e *Engine) SubmitAnswer(c *ws.Client, d ws.AnswerSubmitData) {
	room, ok := e.store.Get(c.Room())
	if !ok {
		c.SendError(ws.ErrRoomNotFound, "xona topilmadi")
		return
	}
	if room.Config.Mode == "time_attack" {
		e.submitTimeAttack(c, room, d)
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
	p := room.Players[userID]
	if p != nil && p.Eliminated {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "o'yindan chiqdingiz")
		return
	}
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

	// Server-authoritative: tur strategiyasi baholaydi, metod strategiyasi ballaydi.
	correct := qtype.For(q.Type).Validate(d.Choice, q.Correct)
	if p != nil {
		modes.For(room.Config.Mode).OnAnswer(p, q, correct, now)
	}
	// Audit: javobni yig'amiz; o'yin tugagach answers_log'ga yoziladi (§7).
	room.Answers = append(room.Answers, state.AnswerEvent{
		UserID: userID, QuestionID: q.ID, Given: d.Choice,
		IsCorrect: correct, TimeMs: int(now - q.AskedAt),
	})
	room.Mu.Unlock()

	// To'g'rilikni OSHKOR QILMAYDI — faqat qabul qilingani.
	c.Send(ws.SAnswerAck, ws.AnswerAckData{Index: d.QuestionIndex, Accepted: true})
}

// submitTimeAttack — per-player oqim: javobni baholaydi va DARHOL keyingi savolni
// shu o'yinchiga yuboradi (reveal yo'q — bu vaqtga poyga). Yagona deadline'gacha.
func (e *Engine) submitTimeAttack(c *ws.Client, room *state.Room, d ws.AnswerSubmitData) {
	userID := c.UserID()

	room.Mu.Lock()
	if room.Status != state.Running {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "o'yin faol emas")
		return
	}
	now := nowMs()
	if now > room.GlobalDeadline {
		room.Mu.Unlock()
		c.SendError(ws.ErrDeadlinePassed, "vaqt tugadi")
		return
	}
	p := room.Players[userID]
	if p == nil || p.TaDone {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "savol qolmadi")
		return
	}
	if d.QuestionIndex != p.TaIdx {
		room.Mu.Unlock()
		c.SendError(ws.ErrBadRequest, "savol indeksi mos emas")
		return
	}
	q := room.Questions[p.TaIdx]
	correct := qtype.For(q.Type).Validate(d.Choice, q.Correct)
	modes.For(room.Config.Mode).OnAnswer(p, q, correct, now)
	room.Answers = append(room.Answers, state.AnswerEvent{
		UserID: userID, QuestionID: q.ID, Given: d.Choice, IsCorrect: correct, TimeMs: 0,
	})
	p.TaIdx++
	total := len(room.Questions)
	idx, deadline := p.TaIdx, room.GlobalDeadline
	var next *state.LiveQuestion
	if p.TaIdx >= total {
		p.TaDone = true
	} else {
		next = room.Questions[p.TaIdx]
	}
	room.Mu.Unlock()

	c.Send(ws.SAnswerAck, ws.AnswerAckData{Index: d.QuestionIndex, Accepted: true})
	if next != nil {
		c.Send(ws.SQuestionShow, e.showPayload(next, idx, total, deadline))
	}
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
	if room.Config.Mode == "time_attack" {
		e.runTimeAttack(room)
		return
	}
	e.runSync(room)
}

// runSync — sinxron oqim (classic/survival/team): hamma bir vaqtda bitta savol.
func (e *Engine) runSync(room *state.Room) {
	sessionID := room.SessionID

	for s := e.Countdown; s > 0; s-- {
		e.hub.BroadcastMsg(sessionID, ws.SGameCountdown, ws.CountdownData{SecondsLeft: s})
		time.Sleep(time.Second)
	}

	room.Mu.RLock()
	total := len(room.Questions)
	timePerQ := room.Config.TimePerQ
	mode := modes.For(room.Config.Mode)
	room.Mu.RUnlock()

	for idx := 0; idx < total; idx++ {
		room.Mu.Lock()
		room.CurrentIdx = idx
		q := room.Questions[idx]
		q.AskedAt = nowMs()
		q.Deadline = q.AskedAt + int64(timePerQ)*1000
		deadline := q.Deadline
		room.Mu.Unlock()

		e.hub.BroadcastMsg(sessionID, ws.SQuestionShow, e.showPayload(q, idx, total, deadline))
		e.scheduleBots(room, q, deadline) // 🏆 botlar deadline ichida javob beradi

		// Deadline'gacha kutish. TODO: hamma javob bersa erta o'tish (early-advance).
		time.Sleep(time.Until(time.UnixMilli(deadline)))

		e.hub.BroadcastMsg(sessionID, ws.SQuestionReveal, ws.QuestionRevealData{
			Index:       idx,
			Correct:     q.Correct, // tur-spetsifik to'g'ri javob (reveal shakli)
			Explanation: q.Explanation,
			Leaderboard: e.leaderboard(room),
			Teams:       e.teamStandings(room), // team rejimida jamoa yig'indisi (aks holda nil)
		})
		if mode.EndEarly(room) { // survival: bittadan kam tirik qolsa
			break
		}
		time.Sleep(e.RevealGap)
	}

	room.Mu.Lock()
	room.Status = state.Finished
	room.Mu.Unlock()

	e.hub.BroadcastMsg(sessionID, ws.SGameOver, ws.GameOverData{
		FinalLeaderboard: e.leaderboard(room),
		Teams:            e.teamStandings(room),
	})

	e.persist(room)

	// Tugagandan keyin biroz turadi (kech reconnect natijani ko'rsin), so'ng tozalanadi.
	time.AfterFunc(60*time.Second, func() { e.store.Delete(sessionID) })
}

// runTimeAttack — per-player oqim: har o'yinchi o'z tezligida, yagona vaqt byudjeti
// ichida iloji boricha ko'p savolga javob beradi. Savol-javob submitTimeAttack'da,
// bu goroutine faqat deadline'gacha (yoki hamma tugaguncha) kutadi va yakunlaydi.
func (e *Engine) runTimeAttack(room *state.Room) {
	sessionID := room.SessionID

	for s := e.Countdown; s > 0; s-- {
		e.hub.BroadcastMsg(sessionID, ws.SGameCountdown, ws.CountdownData{SecondsLeft: s})
		time.Sleep(time.Second)
	}

	room.Mu.Lock()
	total := len(room.Questions)
	budget := int64(room.Config.TimePerQ) * int64(room.Config.QuestionCount) * 1000 // ms
	deadline := nowMs() + budget
	room.GlobalDeadline = deadline
	ids := make([]string, 0, len(room.Players))
	for id, p := range room.Players {
		p.TaIdx, p.TaDone = 0, false
		ids = append(ids, id)
	}
	first := room.Questions[0]
	room.Mu.Unlock()

	// Har o'yinchiga birinchi savol (yagona deadline bilan).
	for _, id := range ids {
		e.hub.SendToUser(sessionID, id, ws.SQuestionShow, e.showPayload(first, 0, total, deadline))
	}

	// Deadline yoki barcha (ulangan) o'yinchilar tugaguncha kutamiz.
	for nowMs() < deadline {
		if e.allTaDone(room) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	room.Mu.Lock()
	room.Status = state.Finished
	room.Mu.Unlock()

	e.hub.BroadcastMsg(sessionID, ws.SGameOver, ws.GameOverData{FinalLeaderboard: e.leaderboard(room)})
	e.persist(room)
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
	for _, a := range room.Answers {
		if !persistentOf[a.UserID] {
			continue // anonim o'yinchi — users FK yo'q
		}
		rec.Answers = append(rec.Answers, AnswerRecord{
			UserID: a.UserID, QuestionID: a.QuestionID, Given: a.Given,
			IsCorrect: a.IsCorrect, TimeMs: a.TimeMs,
		})
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

func (e *Engine) showPayload(q *state.LiveQuestion, idx, total int, deadline int64) ws.QuestionShowData {
	opts := make([]ws.Option, len(q.Options))
	for i, o := range q.Options {
		opts[i] = ws.Option{ID: o.ID, Text: o.Text} // Correct YO'Q
	}
	var targets []ws.Option
	if len(q.Targets) > 0 {
		targets = make([]ws.Option, len(q.Targets))
		for i, tg := range q.Targets {
			targets[i] = ws.Option{ID: tg.ID, Text: tg.Text}
		}
	}
	return ws.QuestionShowData{
		Index: idx, Total: total, Type: q.Type, Prompt: q.Prompt,
		Options: opts, Targets: targets, DeadlineTs: deadline,
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
			UserID: p.UserID, Name: p.Name, Score: p.Score, Connected: p.Connected, IsBot: p.IsBot, Eliminated: p.Eliminated, Team: p.Team,
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
			UserID: p.UserID, Name: p.Name, Score: p.Score, CorrectCnt: p.CorrectCnt, Eliminated: p.Eliminated, Team: p.Team,
		})
	}
	room.Mu.RUnlock()

	// Survival: tirik o'yinchilar yuqorida, keyin ball bo'yicha.
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Eliminated != entries[j].Eliminated {
			return !entries[i].Eliminated
		}
		return entries[i].Score > entries[j].Score
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries
}

// wantsBot — raqib bot/mixed va rejim sinxron (per-player time_attack hozircha botsiz).
func wantsBot(mode, opponent string) bool {
	if opponent != "bot" && opponent != "mixed" {
		return false
	}
	switch mode {
	case "classic", "survival", "team":
		return true
	}
	return false
}

// addBot — xonaga simulyatsion raqib qo'shadi (Persistent emas → DB'ga yozilmaydi).
func addBot(room *state.Room) {
	id := uuid.NewString()
	room.Players[id] = &state.Player{
		UserID: id, Name: "🤖 Bot", Connected: true, IsBot: true, JoinedAt: nowMs(),
	}
}

// scheduleBots — har bot uchun deadline ichida tasodifiy vaqtda ehtimoliy javob
// rejalashtiradi (time.AfterFunc). Bot to'g'ri javobni p ehtimol bilan beradi;
// to'g'ri optionId shart emas — natija (correct bool) bevosita OnAnswer'ga uzatiladi.
func (e *Engine) scheduleBots(room *state.Room, q *state.LiveQuestion, deadline int64) {
	room.Mu.RLock()
	mode := room.Config.Mode
	askedAt := q.AskedAt
	var botIDs []string
	for id, p := range room.Players {
		if p.IsBot && !p.Eliminated {
			botIDs = append(botIDs, id)
		}
	}
	room.Mu.RUnlock()

	span := deadline - askedAt
	if span <= 0 {
		return
	}
	for _, id := range botIDs {
		botID := id
		delay := time.Duration(float64(span)*(0.2+0.6*rand.Float64())) * time.Millisecond
		correct := rand.Float64() < e.BotCorrectProb
		time.AfterFunc(delay, func() {
			room.Mu.Lock()
			defer room.Mu.Unlock()
			if room.Status != state.Running || q.Answered[botID] {
				return
			}
			p := room.Players[botID]
			if p == nil || p.Eliminated {
				return
			}
			q.Answered[botID] = true
			modes.For(mode).OnAnswer(p, q, correct, nowMs())
		})
	}
}

// assignTeams — o'yinchilarni 2 jamoaga balanslab taqsimlaydi (qo'shilish tartibida
// navbatma-navbat A/B). room.Mu LOCK ushlab turilgan holatda chaqiriladi.
func assignTeams(room *state.Room) {
	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return room.Players[ids[i]].JoinedAt < room.Players[ids[j]].JoinedAt
	})
	names := [2]string{"A", "B"}
	for i, id := range ids {
		room.Players[id].Team = names[i%2]
	}
}

// teamStandings — jamoa yig'indisi (faqat team rejimi; aks holda nil → omitempty).
func (e *Engine) teamStandings(room *state.Room) []ws.TeamStanding {
	room.Mu.RLock()
	if room.Config.Mode != "team" {
		room.Mu.RUnlock()
		return nil
	}
	agg := map[string]*ws.TeamStanding{}
	for _, p := range room.Players {
		if p.Team == "" {
			continue
		}
		s := agg[p.Team]
		if s == nil {
			s = &ws.TeamStanding{Team: p.Team}
			agg[p.Team] = s
		}
		s.Score += p.Score
		s.CorrectCnt += p.CorrectCnt
	}
	room.Mu.RUnlock()

	out := make([]ws.TeamStanding, 0, len(agg))
	for _, s := range agg {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

// allTaDone — time_attack: barcha ULANGAN o'yinchilar savollarini tugatdimi.
func (e *Engine) allTaDone(room *state.Room) bool {
	room.Mu.RLock()
	defer room.Mu.RUnlock()
	for _, p := range room.Players {
		if p.Connected && !p.TaDone {
			return false
		}
	}
	return true
}

func buildLive(qs []state.Question) []*state.LiveQuestion {
	live := make([]*state.LiveQuestion, len(qs))
	for i, q := range qs {
		opts := append([]state.Option(nil), q.Options...)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		q.Options = opts
		if len(q.Targets) > 0 { // match/categorize: o'ng tomon/toifalarni ham aralashtiramiz
			tg := append([]state.Option(nil), q.Targets...)
			rand.Shuffle(len(tg), func(a, b int) { tg[a], tg[b] = tg[b], tg[a] })
			q.Targets = tg
		}
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

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
