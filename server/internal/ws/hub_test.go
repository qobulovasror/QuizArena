package ws

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// testServer — httptest + WS handler; ws:// URL qaytaradi.
func testServer(t *testing.T) (string, *Hub) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	srv := httptest.NewServer(http.HandlerFunc(Handle(hub, NewEchoRouter(hub), nil, logger)))
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http"), hub
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func readEnv(t *testing.T, c *websocket.Conn) Envelope {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return env
}

func sendJoin(t *testing.T, c *websocket.Conn, code, name string) {
	t.Helper()
	payload, err := marshalEnvelope(CRoomJoin, RoomJoinData{Code: code, DisplayName: name})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// room:join → room:joined (kimligi) + room:state keladi.
func TestJoinFlow(t *testing.T) {
	url, hub := testServer(t)
	c := dial(t, url)

	sendJoin(t, c, "ABC123", "Ali")

	if env := readEnv(t, c); env.Type != SRoomJoined {
		t.Fatalf("kutilgan %s, keldi %s", SRoomJoined, env.Type)
	}
	if env := readEnv(t, c); env.Type != SRoomState {
		t.Fatalf("kutilgan %s, keldi %s", SRoomState, env.Type)
	}
	if got := hub.RoomSize("ABC123"); got != 1 {
		t.Fatalf("xona hajmi 1 bo'lishi kerak, keldi %d", got)
	}
}

// Ikki klient bir xonada → ikkinchisi qo'shilganda birinchisi room:state oladi.
func TestBroadcastToRoom(t *testing.T) {
	url, _ := testServer(t)
	a := dial(t, url)
	b := dial(t, url)

	sendJoin(t, a, "ROOM1", "A")
	_ = readEnv(t, a) // room:joined
	_ = readEnv(t, a) // room:state (1 o'yinchi)

	sendJoin(t, b, "ROOM1", "B")

	// A ikkinchi o'yinchi qo'shilgani haqida room:state oladi.
	env := readEnv(t, a)
	if env.Type != SRoomState {
		t.Fatalf("A uchun kutilgan %s, keldi %s", SRoomState, env.Type)
	}
	var st RoomStateData
	if err := json.Unmarshal(env.Data, &st); err != nil {
		t.Fatalf("state unmarshal: %v", err)
	}
	if len(st.Players) != 2 {
		t.Fatalf("2 o'yinchi kutilgan, keldi %d", len(st.Players))
	}
}

// Xonaga qo'shilmasdan yuborilgan xabar → BAD_REQUEST.
func TestEchoRequiresRoom(t *testing.T) {
	url, _ := testServer(t)
	c := dial(t, url)

	payload, _ := marshalEnvelope(CGameStart, struct{}{})
	_ = c.WriteMessage(websocket.TextMessage, payload)

	env := readEnv(t, c)
	if env.Type != SError {
		t.Fatalf("kutilgan %s, keldi %s", SError, env.Type)
	}
	var e ErrorData
	_ = json.Unmarshal(env.Data, &e)
	if e.Code != ErrBadRequest {
		t.Fatalf("kutilgan %s, keldi %s", ErrBadRequest, e.Code)
	}
}
