package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/azizbek12234/quizarena/server/internal/auth"
)

type ctxKey int

const userIDKey ctxKey = iota

// requireAuth — JWT tekshiradi va userID'ni kontekstga qo'yadi (REST himoyasi).
func requireAuth(svc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := bearerToken(r)
			if tok == "" {
				writeErr(w, http.StatusUnauthorized, "token kerak")
				return
			}
			claims, err := svc.Verify(tok)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "token yaroqsiz")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requireAdmin — JWT + role=admin tekshiradi (admin marshrutlari).
func requireAdmin(svc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := bearerToken(r)
			if tok == "" {
				writeErr(w, http.StatusUnauthorized, "token kerak")
				return
			}
			claims, err := svc.Verify(tok)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "token yaroqsiz")
				return
			}
			if claims.Role != "admin" {
				writeErr(w, http.StatusForbidden, "admin huquqi kerak")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFrom(r *http.Request) string {
	if v, ok := r.Context().Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return r.URL.Query().Get("token")
}
