// Package httpapi — REST marshrutlari va WebSocket endpoint'i (chi router).
package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/azizbek12234/quizarena/server/internal/auth"
	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/store"
	"github.com/azizbek12234/quizarena/server/internal/ws"
)

// Deps — Router uchun bog'liqliklar.
type Deps struct {
	Cfg      config.Config
	Hub      *ws.Hub
	WSRouter ws.Router
	Auth     *auth.Service
	Queries  *store.Queries
	Logger   *slog.Logger
}

// Router — barcha HTTP/WS marshrutlarini quradi.
func Router(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(requestLogger(d.Logger))
	r.Use(secureHeaders)
	r.Use(corsMiddleware(d.Cfg.CORSOrigins))

	r.Get("/healthz", health)
	r.Get("/ws", ws.Handle(d.Hub, d.WSRouter, wsAuth(d.Auth), d.Logger))

	if d.Auth != nil {
		ah := &authHandler{svc: d.Auth, validate: validator.New(), logger: d.Logger}
		authLimiter := newRateLimiter(30, time.Minute) // brute-force himoyasi (IP/daqiqa)
		r.Route("/api/auth", func(r chi.Router) {
			r.Use(authLimiter.middleware)
			r.Post("/guest", ah.guest)
			r.Post("/register", ah.register)
			r.Post("/login", ah.login)
			r.Post("/telegram", ah.telegram)
		})
	}

	if d.Queries != nil {
		sh := &subjectsHandler{q: d.Queries, logger: d.Logger}
		r.Route("/api/subjects", func(r chi.Router) {
			r.Get("/", sh.list)
			r.Get("/{id}/categories", sh.categories)
		})

		// Global reyting (ochiq, auth shart emas)
		lh := &leaderboardHandler{q: d.Queries, logger: d.Logger}
		r.Get("/api/leaderboard/global", lh.global)
	}

	if d.Auth != nil && d.Queries != nil {
		mh := &meHandler{q: d.Queries, logger: d.Logger}
		sh := &srsHandler{q: d.Queries, validate: validator.New(), logger: d.Logger}
		ah := &assessHandler{q: d.Queries, logger: d.Logger}
		r.Group(func(r chi.Router) {
			r.Use(requireAuth(d.Auth))
			r.Get("/api/me/history", mh.history)
			r.Get("/api/me/rating", mh.rating)   // 🏆 1v1 ELO reyting
			r.Get("/api/me/srs/due", sh.due)     // 📚 takror kartalar
			r.Post("/api/srs/review", sh.review) // 📚 baho → SM-2
			r.Get("/api/me/assessment", ah.questions)
			r.Post("/api/me/assessment/submit", ah.submit) // 📊 baholash → mastery
			r.Get("/api/me/mastery", ah.mastery)
		})

		// Admin (RBAC: role=admin)
		adm := &adminHandler{q: d.Queries, validate: validator.New(), logger: d.Logger}
		r.Group(func(r chi.Router) {
			r.Use(requireAdmin(d.Auth))
			r.Post("/api/admin/subjects", adm.createSubject)
			r.Post("/api/admin/categories", adm.createCategory)
			r.Post("/api/admin/questions", adm.createQuestion)
			r.Get("/api/admin/questions", adm.listQuestions)
			r.Delete("/api/admin/questions/{id}", adm.deleteQuestion)
		})
	}

	return r
}

// wsAuth — WS ulanishida tokendan userID chiqaradi (mehmon/akkaunt). Token bo'lmasa anonim.
func wsAuth(svc *auth.Service) ws.AuthFunc {
	if svc == nil {
		return nil
	}
	return func(r *http.Request) (string, bool) {
		tok := r.URL.Query().Get("token")
		if tok == "" {
			tok = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if tok == "" {
			return "", false
		}
		claims, err := svc.Verify(tok)
		if err != nil {
			return "", false
		}
		return claims.Subject, true
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
