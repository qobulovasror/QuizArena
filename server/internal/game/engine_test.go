package game

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
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
	e := NewEngine(ws.NewHub(logger), store, providers.NewSample(), logger)
	e.Countdown = 0 // testda sanoqsiz
	e.RevealGap = 10 * time.Millisecond
	return e, store
}

func dialWS(t *testing.T, e *Engine) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(ws.Handle(e.hub, NewRouter(e), nil, e.logger))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
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
	room, ok := store.Get(rj.SessionID)
	if !ok {
		t.Fatal("xona topilmadi")
	}
	room.Mu.RLock()
	correctID := room.Questions[0].CorrectID()
	room.Mu.RUnlock()

	send(t, conn, ws.CAnswerSubmit, ws.AnswerSubmitData{
		QuestionIndex: 0, Choice: mustJSON(map[string]string{"optionId": correctID}),
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

// Ball: to'g'ri javob 100..200; tezroq = ko'proq.
func TestScoreFor(t *testing.T) {
	q := &state.LiveQuestion{AskedAt: 0, Deadline: 1000}
	fast := scoreFor(q, 0)   // darhol
	slow := scoreFor(q, 900) // deyarli deadline
	if fast < 199 || fast > 201 {
		t.Fatalf("tez javob ~200 kutilgan, %v", fast)
	}
	if slow <= 100 || slow >= fast {
		t.Fatalf("sekin javob 100<slow<fast kutilgan, %v", slow)
	}
}
