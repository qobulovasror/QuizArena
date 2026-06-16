package game

import (
	"encoding/json"

	"github.com/azizbek12234/quizarena/server/internal/ws"
)

// Router — WebSocket xabarlarini engine amallariga bog'laydi (ws.Router interfeysi).
// EchoRouter o'rnini bosadi.
type Router struct{ e *Engine }

func NewRouter(e *Engine) *Router { return &Router{e: e} }

func (r *Router) OnConnect(c *ws.Client) {}

func (r *Router) OnDisconnect(c *ws.Client) {
	r.e.HandleDisconnect(c)
}

func (r *Router) Route(c *ws.Client, env ws.Envelope) {
	switch env.Type {
	case ws.CRoomCreate:
		var d ws.RoomCreateData
		if !decode(c, env.Data, &d) {
			return
		}
		r.e.CreateRoom(c, d)

	case ws.CRoomJoin:
		var d ws.RoomJoinData
		if !decode(c, env.Data, &d) {
			return
		}
		r.e.JoinRoom(c, d)

	case ws.CRoomResume:
		var d ws.RoomResumeData
		if !decode(c, env.Data, &d) {
			return
		}
		r.e.Resume(c, d)

	case ws.CRoomLeave:
		r.e.Leave(c)

	case ws.CGameStart:
		r.e.StartGame(c)

	case ws.CAnswerSubmit:
		var d ws.AnswerSubmitData
		if !decode(c, env.Data, &d) {
			return
		}
		r.e.SubmitAnswer(c, d)

	default:
		c.SendError(ws.ErrInvalidMessage, "noma'lum xabar turi")
	}
}

func decode(c *ws.Client, raw json.RawMessage, v any) bool {
	if err := json.Unmarshal(raw, v); err != nil {
		c.SendError(ws.ErrBadRequest, "yuk yaroqsiz")
		return false
	}
	return true
}
