package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// DEV: barcha origin'larga ruxsat. PROD: CORS ro'yxati bilan cheklash (PLAN.md §11 B6).
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Router — kiruvchi xabarlarni boshqaradi. Transportni o'yin mantig'idan ajratadi.
// Bosqich 1 da EchoRouter; keyin GameRouter (engine) bilan almashtiriladi.
type Router interface {
	OnConnect(c *Client)
	Route(c *Client, env Envelope)
	OnDisconnect(c *Client)
}

// AuthFunc — ulanish so'rovidan autentifikatsiyalangan userID chiqaradi (token).
// nil bo'lsa yoki ok=false bo'lsa, ulanish anonim (engine guest id beradi).
type AuthFunc func(r *http.Request) (userID string, ok bool)

// Handle — HTTP ulanishini WebSocket'ga ko'taradi va klientni ishga tushiradi.
func Handle(hub *Hub, router Router, authFn AuthFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authUserID string
		if authFn != nil {
			if uid, ok := authFn(r); ok {
				authUserID = uid
			}
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("ws: upgrade muvaffaqiyatsiz", "err", err)
			return
		}
		c := &Client{
			hub:        hub,
			conn:       conn,
			send:       make(chan []byte, sendBuffer),
			router:     router,
			logger:     logger,
			id:         uuid.NewString(),
			authUserID: authUserID,
		}
		logger.Debug("ws: ulanish ochildi", "client", c.id)
		if router != nil {
			router.OnConnect(c)
		}
		go c.writePump()
		go c.readPump()
	}
}

// EchoRouter — Bosqich 1 vaqtinchalik router: transport ishlayotganini tekshirish uchun.
//
//   - room:join → xonaga qo'shilish, room:joined (kimligi) + room:state broadcast
//   - room:leave → chiqish
//   - qolganlari → xonaga echo (agar xonada bo'lsa)
//
// To'liq o'yin mantig'i (create/start/answer + ball + deadline) keyingi qadamda
// game engine orqali ulanadi va bu router GameRouter bilan almashtiriladi.
type EchoRouter struct{ hub *Hub }

func NewEchoRouter(hub *Hub) *EchoRouter { return &EchoRouter{hub: hub} }

func (r *EchoRouter) OnConnect(c *Client) {}

func (r *EchoRouter) OnDisconnect(c *Client) {
	if c.room != "" {
		r.broadcastState(c.room)
	}
}

func (r *EchoRouter) Route(c *Client, env Envelope) {
	switch env.Type {
	case CRoomJoin:
		var d RoomJoinData
		if err := json.Unmarshal(env.Data, &d); err != nil || d.Code == "" {
			c.SendError(ErrBadRequest, "code va displayName kerak")
			return
		}
		c.name = d.DisplayName
		r.hub.Join(c, d.Code)
		c.Send(SRoomJoined, RoomJoinedData{
			SessionID:   d.Code,
			UserID:      c.id,
			ResumeToken: c.id, // vaqtinchalik; haqiqiy token keyin (auth)
		})
		r.broadcastState(d.Code)

	case CRoomLeave:
		room := c.room
		r.hub.Leave(c)
		if room != "" {
			r.broadcastState(room)
		}

	default:
		// Vaqtinchalik echo (transportni tekshirish uchun).
		if c.room == "" {
			c.SendError(ErrBadRequest, "avval room:join yuboring")
			return
		}
		if raw, err := marshalEnvelope(env.Type, json.RawMessage(env.Data)); err == nil {
			r.hub.Broadcast(c.room, raw)
		}
	}
}

func (r *EchoRouter) broadcastState(room string) {
	payload, err := marshalEnvelope(SRoomState, RoomStateData{
		SessionID: room,
		Code:      room,
		Status:    "lobby",
		Players:   r.hub.Players(room),
	})
	if err != nil {
		return
	}
	r.hub.Broadcast(room, payload)
}
