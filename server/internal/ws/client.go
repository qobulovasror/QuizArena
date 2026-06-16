package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // bitta yozuvga ruxsat etilgan vaqt
	pongWait       = 60 * time.Second    // klientdan pong kutish muddati
	pingPeriod     = (pongWait * 9) / 10 // ping yuborish davri (< pongWait)
	maxMessageSize = 1 << 20             // 1 MB
	sendBuffer     = 32                  // chiquvchi navbat hajmi
)

// Client — bitta WebSocket ulanishi.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	router Router
	logger *slog.Logger

	id         string // ulanish identifikatori (har socket uchun)
	authUserID string // tokendan kelgan userID (bo'lsa); engine shuni afzal ko'radi
	userID     string // mantiqiy o'yinchi identifikatori (auth yoki guest)
	name       string // displayName (xona ichida)
	room       string // joriy xona (faqat hub.mu ostida o'zgaradi)
}

// AuthUserID — token orqali autentifikatsiyalangan userID ("" agar anonim).
func (c *Client) AuthUserID() string { return c.authUserID }

// SetIdentity — o'yinchi kimligini o'rnatadi (create/join/resume'da engine chaqiradi).
func (c *Client) SetIdentity(userID, name string) {
	c.userID = userID
	c.name = name
}

// UserID — mantiqiy o'yinchi id'si.
func (c *Client) UserID() string { return c.userID }

// Name — displayName.
func (c *Client) Name() string { return c.name }

// Send — tip va yukni konvertga o'rab, chiquvchi navbatga qo'yadi.
func (c *Client) Send(t MsgType, data any) {
	payload, err := marshalEnvelope(t, data)
	if err != nil {
		c.logger.Error("ws: konvert marshal", "err", err, "type", t)
		return
	}
	select {
	case c.send <- payload:
	default:
		c.logger.Warn("ws: chiquvchi navbat to'la, xabar tashlandi", "client", c.id, "type", t)
	}
}

// SendError — standart xato xabari.
func (c *Client) SendError(code, message string) {
	c.Send(SError, ErrorData{Code: code, Message: message})
}

// ID / Room / Name — o'qish uchun yordamchilar.
func (c *Client) ID() string   { return c.id }
func (c *Client) Room() string { return c.room }

// readPump — ulanishdan xabarlarni o'qiydi va router'ga uzatadi.
// Bitta ulanish uchun bitta goroutine'da ishlaydi.
func (c *Client) readPump() {
	defer func() {
		c.hub.Leave(c)
		if c.router != nil {
			c.router.OnDisconnect(c)
		}
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.logger.Debug("ws: o'qish uzildi", "client", c.id, "err", err)
			}
			break
		}
		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			c.SendError(ErrInvalidMessage, "JSON buzuq")
			continue
		}
		if c.router != nil {
			c.router.Route(c, env)
		}
	}
}

// writePump — chiquvchi navbat va ping'larni ulanishga yozadi.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // navbat yopildi
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// marshalEnvelope — { type, data } konvertini baytlarga aylantiradi.
func marshalEnvelope(t MsgType, data any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: t, Data: raw})
}
