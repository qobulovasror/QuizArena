package game

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/azizbek12234/quizarena/server/internal/game/providers"
	"github.com/azizbek12234/quizarena/server/internal/state"
	"github.com/azizbek12234/quizarena/server/internal/ws"
)

func newTestEngine() (*Engine, *state.MemStore) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := state.NewMemStore()
	reg := NewRegistry(providers.NewSample())
	if ev, err := providers.NewEnglishVerb(); err == nil {
		reg.Register("english", ev)
	}
	e := NewEngine(ws.NewHub(logger), store, reg, nil, logger)
	e.Countdown = 0 // testda sanoqsiz
	e.RevealGap = 10 * time.Millisecond
	return e, store
}

func dialWS(t *testing.T, e *Engine) *websocket.Conn {
	return dialWSWith(t, e, nil)
}

func dialWSWith(t *testing.T, e *Engine, authFn ws.AuthFunc) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(ws.Handle(e.hub, NewRouter(e), authFn, e.logger))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// dialShared — bitta server (ya'ni bitta Router/Matchmaker) ga n ulanish.
// Matchmaking testi uchun kerak: productionda Router bitta, harness esa har
// dialWS uchun yangi Router yaratadi (umumiy hub/store orqali xona ishlaydi,
// lekin matchmaker holati Router ichida).
func dialShared(t *testing.T, e *Engine, n int) []*websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(ws.Handle(e.hub, NewRouter(e), nil, e.logger))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conns := make([]*websocket.Conn, n)
	for i := range conns {
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		t.Cleanup(func() { _ = c.Close() })
		conns[i] = c
	}
	return conns
}

func jraw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

// correctOptionID — server holatidan birinchi savolning to'g'ri optionId'sini oladi (oq-quti).
func correctOptionID(t *testing.T, store *state.MemStore, sessionID string) string {
	return correctOptionIDAt(t, store, sessionID, 0)
}

// correctOptionIDAt — idx'inchi savolning to'g'ri optionId'si (time_attack uchun).
func correctOptionIDAt(t *testing.T, store *state.MemStore, sessionID string, idx int) string {
	t.Helper()
	room, ok := store.Get(sessionID)
	if !ok {
		t.Fatal("xona topilmadi")
	}
	room.Mu.RLock()
	raw := room.Questions[idx].Correct
	room.Mu.RUnlock()
	var cc struct {
		OptionID string `json:"optionId"`
	}
	_ = json.Unmarshal(raw, &cc)
	return cc.OptionID
}

func send(t *testing.T, conn *websocket.Conn, typ ws.MsgType, data any) {
	t.Helper()
	body, _ := json.Marshal(data)
	env, _ := json.Marshal(ws.Envelope{Type: typ, Data: body})
	if err := conn.WriteMessage(websocket.TextMessage, env); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func expect(t *testing.T, conn *websocket.Conn, typ ws.MsgType) ws.Envelope {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read (%s kutilgan): %v", typ, err)
	}
	var env ws.Envelope
	_ = json.Unmarshal(raw, &env)
	if env.Type != typ {
		t.Fatalf("kutilgan %s, keldi %s", typ, env.Type)
	}
	return env
}

// To'liq classic o'yin: create → start → question → answer → reveal → over.
func TestFullGameFlow(t *testing.T) {
	e, store := newTestEngine()
	conn := dialWS(t, e)

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "x", Mode: "classic", QuestionCount: 1, TimePerQ: 1, DisplayName: "Host",
	})
	joined := expect(t, conn, ws.SRoomJoined)
	var rj ws.RoomJoinedData
	_ = json.Unmarshal(joined.Data, &rj)
	expect(t, conn, ws.SRoomState)

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // status=running

	show := expect(t, conn, ws.SQuestionShow)
	var qs ws.QuestionShowData
	_ = json.Unmarshal(show.Data, &qs)
	if len(qs.Options) == 0 {
		t.Fatal("variantlar bo'sh")
	}
	// Anti-cheat: question:show ichida to'g'ri javob bo'lmasligi kerak (tip darajasida ham yo'q).
	if strings.Contains(string(show.Data), "correct") {
		t.Fatal("question:show to'g'ri javobni oshkor qildi")
	}

	// To'g'ri variant id'sini server holatidan olamiz (oq-quti tekshiruv).
	correctID := correctOptionID(t, store, rj.SessionID)
	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{
		QuestionIndex: 0, Choice: jraw(map[string]string{"optionId": correctID}),
	})
	expect(t, conn, ws.SAnswerAck)

	reveal := expect(t, conn, ws.SQuestionReveal)
	var rv ws.QuestionRevealData
	_ = json.Unmarshal(reveal.Data, &rv)
	if len(rv.Leaderboard) != 1 || rv.Leaderboard[0].Score <= 0 {
		t.Fatalf("to'g'ri javobdan keyin ball > 0 kutilgan: %+v", rv.Leaderboard)
	}

	over := expect(t, conn, ws.SGameOver)
	var ov ws.GameOverData
	_ = json.Unmarshal(over.Data, &ov)
	if len(ov.FinalLeaderboard) != 1 || ov.FinalLeaderboard[0].CorrectCnt != 1 {
		t.Fatalf("correctCnt 1 kutilgan: %+v", ov.FinalLeaderboard)
	}
}

// time_attack: per-player oqim — javobdan keyin DARHOL keyingi savol (reveal yo'q),
// hamma tugagach game:over.
func TestTimeAttackFlow(t *testing.T) {
	e, store := newTestEngine()
	conn := dialWS(t, e)

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "x", Mode: "time_attack", QuestionCount: 2, TimePerQ: 1, DisplayName: "Host",
	})
	joined := expect(t, conn, ws.SRoomJoined)
	var rj ws.RoomJoinedData
	_ = json.Unmarshal(joined.Data, &rj)
	expect(t, conn, ws.SRoomState)

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // running

	show0 := expect(t, conn, ws.SQuestionShow)
	var q0 ws.QuestionShowData
	_ = json.Unmarshal(show0.Data, &q0)
	if q0.Index != 0 {
		t.Fatalf("birinchi savol index 0 kutilgan: %d", q0.Index)
	}

	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{
		QuestionIndex: 0, Choice: jraw(map[string]string{"optionId": correctOptionIDAt(t, store, rj.SessionID, 0)}),
	})
	expect(t, conn, ws.SAnswerAck)
	show1 := expect(t, conn, ws.SQuestionShow) // reveal'siz darhol keyingi savol
	var q1 ws.QuestionShowData
	_ = json.Unmarshal(show1.Data, &q1)
	if q1.Index != 1 {
		t.Fatalf("keyingi savol index 1 kutilgan: %d", q1.Index)
	}

	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{
		QuestionIndex: 1, Choice: jraw(map[string]string{"optionId": correctOptionIDAt(t, store, rj.SessionID, 1)}),
	})
	expect(t, conn, ws.SAnswerAck)

	over := expect(t, conn, ws.SGameOver) // hamma tugadi → erta yakun
	var ov ws.GameOverData
	_ = json.Unmarshal(over.Data, &ov)
	if len(ov.FinalLeaderboard) != 1 || ov.FinalLeaderboard[0].CorrectCnt != 2 {
		t.Fatalf("2 to'g'ri javob kutilgan: %+v", ov.FinalLeaderboard)
	}
}

// teamStandings: jamoa yig'indisi to'g'ri hisoblanadi va saralanadi.
func TestTeamStandings(t *testing.T) {
	e, _ := newTestEngine()
	room := &state.Room{
		Config: state.Config{Mode: "team"},
		Players: map[string]*state.Player{
			"a": {Team: "A", Score: 100, CorrectCnt: 1},
			"b": {Team: "B", Score: 250, CorrectCnt: 2},
			"c": {Team: "A", Score: 200, CorrectCnt: 2},
		},
	}
	ts := e.teamStandings(room)
	if len(ts) != 2 {
		t.Fatalf("2 jamoa kutilgan: %+v", ts)
	}
	if ts[0].Team != "A" || ts[0].Score != 300 || ts[0].Rank != 1 { // A: 100+200=300 > B: 250
		t.Fatalf("A jamoa 300 ball rank 1 kutilgan: %+v", ts[0])
	}
	room.Config.Mode = "classic"
	if e.teamStandings(room) != nil {
		t.Fatal("team bo'lmagan rejimda nil kutilgan")
	}
}

// 🏆 Matchmaking: ikki o'yinchi navbatga qo'shilsa juftlanadi va duel boshlanadi.
func TestMatchmakingDuel(t *testing.T) {
	e, _ := newTestEngine()
	conns := dialShared(t, e, 2)
	a, b := conns[0], conns[1]

	send(t, a, ws.CMatchQueue, ws.MatchQueueData{SubjectID: "x", DisplayName: "A"})
	expect(t, a, ws.SMatchQueued) // A kutmoqda

	send(t, b, ws.CMatchQueue, ws.MatchQueueData{SubjectID: "x", DisplayName: "B"})

	// Ikkalasi ham raqib topdi → duel boshlanadi.
	for _, c := range []*websocket.Conn{a, b} {
		found := expect(t, c, ws.SMatchFound)
		var mf ws.MatchFoundData
		_ = json.Unmarshal(found.Data, &mf)
		if mf.VsBot {
			t.Fatal("inson-inson duel kutilgan (VsBot=false)")
		}
		expect(t, c, ws.SRoomJoined)
		st := expect(t, c, ws.SRoomState)
		var rs ws.RoomStateData
		_ = json.Unmarshal(st.Data, &rs)
		if len(rs.Players) != 2 || rs.Status != "running" {
			t.Fatalf("2 o'yinchili running duel kutilgan: %+v", rs)
		}
		expect(t, c, ws.SQuestionShow) // o'yin ketmoqda
	}
}

// botProb — qiyilik darajasi ehtimolga to'g'ri xaritalanadi; bo'sh → BotCorrectProb.
func TestBotProb(t *testing.T) {
	e, _ := newTestEngine()
	mk := func(d string) *state.Room { return &state.Room{Config: state.Config{BotDifficulty: d}} }
	if e.botProb(mk("easy")) != 0.45 || e.botProb(mk("medium")) != 0.65 || e.botProb(mk("hard")) != 0.85 {
		t.Fatal("qiyinlik → ehtimol mapping xato")
	}
	e.BotCorrectProb = 0.7
	if e.botProb(mk("")) != 0.7 {
		t.Fatalf("default BotCorrectProb kutilgan: %v", e.botProb(mk("")))
	}
}

// 🏆 time_attack'da bot: xonaga bot qo'shiladi va o'yin yakunida leaderboard'da bo'ladi.
func TestTimeAttackBot(t *testing.T) {
	e, store := newTestEngine()
	e.BotCorrectProb = 1.0
	conn := dialWS(t, e)

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "x", Mode: "time_attack", Opponent: "bot", QuestionCount: 2, TimePerQ: 1, DisplayName: "Host",
	})
	joined := expect(t, conn, ws.SRoomJoined)
	var rj ws.RoomJoinedData
	_ = json.Unmarshal(joined.Data, &rj)
	st := expect(t, conn, ws.SRoomState)
	var rs ws.RoomStateData
	_ = json.Unmarshal(st.Data, &rs)
	if len(rs.Players) != 2 {
		t.Fatalf("time_attack'da host+bot kutilgan: %+v", rs.Players)
	}

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // running
	expect(t, conn, ws.SQuestionShow)
	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{QuestionIndex: 0, Choice: jraw(map[string]string{"optionId": correctOptionIDAt(t, store, rj.SessionID, 0)})})
	expect(t, conn, ws.SAnswerAck)
	expect(t, conn, ws.SQuestionShow)
	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{QuestionIndex: 1, Choice: jraw(map[string]string{"optionId": correctOptionIDAt(t, store, rj.SessionID, 1)})})
	expect(t, conn, ws.SAnswerAck)

	over := expect(t, conn, ws.SGameOver)
	var ov ws.GameOverData
	_ = json.Unmarshal(over.Data, &ov)
	if len(ov.FinalLeaderboard) != 2 {
		t.Fatalf("2 o'yinchi (host+bot) kutilgan: %+v", ov.FinalLeaderboard)
	}
}

// 🏆 Bot raqib: opponent=bot xonada bot o'yinchi paydo bo'ladi va javob beradi.
func TestBotOpponent(t *testing.T) {
	e, _ := newTestEngine()
	e.BotCorrectProb = 1.0 // deterministik: bot har doim to'g'ri
	conn := dialWS(t, e)

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "x", Mode: "classic", Opponent: "bot", QuestionCount: 1, TimePerQ: 1, DisplayName: "Host",
	})
	expect(t, conn, ws.SRoomJoined)
	st := expect(t, conn, ws.SRoomState)
	var rs ws.RoomStateData
	_ = json.Unmarshal(st.Data, &rs)
	if len(rs.Players) != 2 { // host + bot
		t.Fatalf("lobby'da 2 o'yinchi (host+bot) kutilgan: %+v", rs.Players)
	}

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // running
	expect(t, conn, ws.SQuestionShow)
	// Host javob bermaydi; bot deadline ichida javob beradi.
	expect(t, conn, ws.SQuestionReveal)
	over := expect(t, conn, ws.SGameOver)
	var ov ws.GameOverData
	_ = json.Unmarshal(over.Data, &ov)

	var bot *ws.LeaderboardEntry
	for i := range ov.FinalLeaderboard {
		if ov.FinalLeaderboard[i].Name == "🤖 Bot" {
			bot = &ov.FinalLeaderboard[i]
		}
	}
	if bot == nil || bot.CorrectCnt != 1 {
		t.Fatalf("bot 1 to'g'ri javob bilan kutilgan: %+v", ov.FinalLeaderboard)
	}
}

// match/categorize: targets (o'ng tomon/toifalar) question:show payload'iga kiradi.
func TestShowPayloadTargets(t *testing.T) {
	e, _ := newTestEngine()
	q := &state.LiveQuestion{Question: state.Question{
		Type:    "match",
		Prompt:  "moslang",
		Options: []state.Option{{ID: "l1", Text: "cat"}},
		Targets: []state.Option{{ID: "r1", Text: "mushuk"}},
	}}
	p := e.showPayload(q, 0, 1, 123)
	if len(p.Targets) != 1 || p.Targets[0].ID != "r1" {
		t.Fatalf("targets showPayload'da kutilgan: %+v", p.Targets)
	}
	if len(p.Options) != 1 || p.Options[0].ID != "l1" {
		t.Fatalf("options showPayload'da kutilgan: %+v", p.Options)
	}
}

// "english" sohasida EnglishVerbProvider'dan haqiqiy savol keladi.
func TestEnglishGameQuestion(t *testing.T) {
	e, _ := newTestEngine()
	conn := dialWS(t, e)

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "english", Mode: "classic", QuestionCount: 1, TimePerQ: 1, DisplayName: "Host",
	})
	expect(t, conn, ws.SRoomJoined)
	expect(t, conn, ws.SRoomState)

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // status=running
	show := expect(t, conn, ws.SQuestionShow)
	var qs ws.QuestionShowData
	_ = json.Unmarshal(show.Data, &qs)

	if !strings.Contains(qs.Prompt, "fe'l") {
		t.Fatalf("ingliz fe'l savoli kutilgan, keldi: %q", qs.Prompt)
	}
	if len(qs.Options) != 4 {
		t.Fatalf("4 variant kutilgan, %d", len(qs.Options))
	}
}

// Faqat host o'yinni boshlay oladi.
func TestOnlyHostCanStart(t *testing.T) {
	e, _ := newTestEngine()
	host := dialWS(t, e)
	send(t, host, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "x", Mode: "classic", QuestionCount: 1, TimePerQ: 1, DisplayName: "Host",
	})
	expect(t, host, ws.SRoomJoined)
	state0 := expect(t, host, ws.SRoomState)
	var st ws.RoomStateData
	_ = json.Unmarshal(state0.Data, &st)

	// Ikkinchi o'yinchi qo'shiladi va start'ga urinadi → NOT_HOST.
	guest := dialWS(t, e)
	send(t, guest, ws.CRoomJoin, ws.RoomJoinData{Code: st.Code, DisplayName: "Guest"})
	expect(t, guest, ws.SRoomJoined)
	expect(t, guest, ws.SRoomState)

	send(t, guest, ws.CGameStart, struct{}{})
	errEnv := expect(t, guest, ws.SError)
	var ed ws.ErrorData
	_ = json.Unmarshal(errEnv.Data, &ed)
	if ed.Code != ws.ErrNotHost {
		t.Fatalf("kutilgan %s, keldi %s", ws.ErrNotHost, ed.Code)
	}
}

// fakePersister — SaveGame chaqiruvini ushlaydi.
type fakePersister struct {
	mu   sync.Mutex
	recs []GameRecord
}

func (f *fakePersister) SaveGame(_ context.Context, rec GameRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recs = append(f.recs, rec)
	return nil
}

func (f *fakePersister) last() (GameRecord, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.recs) == 0 {
		return GameRecord{}, false
	}
	return f.recs[len(f.recs)-1], true
}

// Tokenli (persistent) host o'yini tugagach SaveGame chaqiriladi.
func TestPersistOnGameEnd(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := state.NewMemStore()
	fp := &fakePersister{}
	e := NewEngine(ws.NewHub(logger), store, NewRegistry(providers.NewSample()), fp, logger)
	e.Countdown = 0
	e.RevealGap = 10 * time.Millisecond

	const hostID = "11111111-1111-1111-1111-111111111111"
	conn := dialWSWith(t, e, func(*http.Request) (string, bool) { return hostID, true })

	send(t, conn, ws.CRoomCreate, ws.RoomCreateData{
		SubjectID: "english", Mode: "classic", QuestionCount: 1, TimePerQ: 1, DisplayName: "Host",
	})
	joined := expect(t, conn, ws.SRoomJoined)
	var rj ws.RoomJoinedData
	_ = json.Unmarshal(joined.Data, &rj)
	if rj.UserID != hostID {
		t.Fatalf("tokenli userID kutilgan %s, keldi %s", hostID, rj.UserID)
	}
	expect(t, conn, ws.SRoomState)

	send(t, conn, ws.CGameStart, struct{}{})
	expect(t, conn, ws.SRoomState) // status=running
	expect(t, conn, ws.SQuestionShow)
	correctID := correctOptionID(t, store, rj.SessionID)
	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{
		QuestionIndex: 0, Choice: jraw(map[string]string{"optionId": correctID}),
	})
	expect(t, conn, ws.SAnswerAck)
	expect(t, conn, ws.SQuestionReveal)
	expect(t, conn, ws.SGameOver)

	// persist goroutine'i tugashini biroz kutamiz.
	var rec GameRecord
	for i := 0; i < 50; i++ {
		if r, ok := fp.last(); ok {
			rec = r
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if rec.HostUserID != hostID {
		t.Fatalf("SaveGame host %s kutilgan, keldi %q", hostID, rec.HostUserID)
	}
	if len(rec.Results) != 1 || rec.Results[0].UserID != hostID || rec.Results[0].Rank != 1 {
		t.Fatalf("1 natija (rank 1) kutilgan: %+v", rec.Results)
	}
	if rec.Results[0].CorrectCnt != 1 {
		t.Fatalf("correctCnt 1 kutilgan: %+v", rec.Results[0])
	}
}
