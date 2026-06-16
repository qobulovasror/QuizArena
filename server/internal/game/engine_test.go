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

func jraw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

// correctOptionID — server holatidan birinchi savolning to'g'ri optionId'sini oladi (oq-quti).
func correctOptionID(t *testing.T, store *state.MemStore, sessionID string) string {
	t.Helper()
	room, ok := store.Get(sessionID)
	if !ok {
		t.Fatal("xona topilmadi")
	}
	room.Mu.RLock()
	raw := room.Questions[0].Correct
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
