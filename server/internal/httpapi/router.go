// Package httpapi — REST marshrutlari va WebSocket endpoint'i (chi router).
package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

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
	r.Use(cors)

	r.Get("/healthz", health)
	r.Get("/ws", ws.Handle(d.Hub, d.WSRouter, wsAuth(d.Auth), d.Logger))

	if d.Auth != nil {
		ah := &authHandler{svc: d.Auth, validate: validator.New(), logger: d.Logger}
		r.Route("/api/auth", func(r chi.Router) {
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
	}

	if d.Auth != nil && d.Queries != nil {
		mh := &meHandler{q: d.Queries, logger: d.Logger}
		r.Route("/api/me", func(r chi.Router) {
			r.Use(requireAuth(d.Auth))
			r.Get("/history", mh.history)
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

// cors — DEV uchun ochiq CORS. PROD'da origin ro'yxati bilan cheklanadi.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
