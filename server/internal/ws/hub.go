package ws

import (
	"log/slog"
	"sync"
)

// Hub — barcha faol ulanishlar va xonalarni boshqaradi (bitta instans).
// Goroutine-xavfsiz: xonalar xaritasi RWMutex bilan himoyalangan.
//
// Scaling bosqichida (PLAN.md §11) bu qatlam Redis pub/sub bilan almashtiriladi —
// interfeys o'zgarmaydi.
type Hub struct {
	mu     sync.RWMutex
	rooms  map[string]map[*Client]struct{} // roomID -> clientlar
	logger *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:  make(map[string]map[*Client]struct{}),
		logger: logger,
	}
}

// Join — klientni xonaga qo'shadi (avvalgi xonadan chiqaradi).
func (h *Hub) Join(c *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeLocked(c)
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]struct{})
	}
	h.rooms[room][c] = struct{}{}
	c.room = room
}

// Leave — klientni joriy xonasidan chiqaradi.
func (h *Hub) Leave(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeLocked(c)
}

func (h *Hub) removeLocked(c *Client) {
	if c.room == "" {
		return
	}
	if set := h.rooms[c.room]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.rooms, c.room)
		}
	}
	c.room = ""
}

// Broadcast — xonadagi barcha klientlarga xom xabar yuboradi.
// Sekin iste'molchi (to'la bufer) o'tkazib yuboriladi (keyin uzib tashlanadi).
func (h *Hub) Broadcast(room string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[room] {
		select {
		case c.send <- payload:
		default:
			h.logger.Warn("ws: sekin iste'molchi, xabar tashlandi", "client", c.id, "room", room)
		}
	}
}

// BroadcastMsg — tip+yukni konvertga o'rab, xonaga broadcast qiladi.
func (h *Hub) BroadcastMsg(room string, t MsgType, data any) {
	payload, err := marshalEnvelope(t, data)
	if err != nil {
		h.logger.Error("ws: broadcast marshal", "err", err, "type", t)
		return
	}
	h.Broadcast(room, payload)
}

// SendToUser — xonadagi muayyan userID'li klient(lar)ga xabar yuboradi.
// time_attack per-player oqimi uchun (har o'yinchiga o'z savoli).
func (h *Hub) SendToUser(room, userID string, t MsgType, data any) {
	payload, err := marshalEnvelope(t, data)
	if err != nil {
		h.logger.Error("ws: sendToUser marshal", "err", err, "type", t)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[room] {
		if c.userID != userID {
			continue
		}
		select {
		case c.send <- payload:
		default:
			h.logger.Warn("ws: sekin iste'molchi, xabar tashlandi", "client", c.id, "room", room)
		}
	}
}

// Players — xonadagi klientlarning snapshot ro'yxati.
func (h *Hub) Players(room string) []Player {
	h.mu.RLock()
	defer h.mu.RUnlock()
	players := make([]Player, 0, len(h.rooms[room]))
	for c := range h.rooms[room] {
		players = append(players, Player{
			UserID:    c.id,
			Name:      c.name,
			Connected: true,
		})
	}
	return players
}

// RoomSize — xonadagi klientlar soni.
func (h *Hub) RoomSize(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[room])
}
