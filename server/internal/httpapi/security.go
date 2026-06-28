package httpapi

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// corsMiddleware — origins bo'sh bo'lsa hammaga ruxsat (dev); aks holda ro'yxat.
func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	allowAll := len(origins) == 0
	set := make(map[string]bool, len(origins))
	for _, o := range origins {
		set[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			switch {
			case allowAll:
				w.Header().Set("Access-Control-Allow-Origin", "*")
			case origin != "" && set[origin]:
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// secureHeaders — asosiy xavfsizlik sarlavhalari (API javoblari uchun).
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// requestLogger — har so'rovni slog bilan loglaydi (metod, yo'l, status, davomiylik).
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			logger.Info("http",
				"method", r.Method, "path", r.URL.Path,
				"status", sw.status, "ms", time.Since(start).Milliseconds())
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Hijack — WebSocket upgrade (gorilla) ishlashi uchun delegatsiya qilamiz.
func (s *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack qo'llab-quvvatlanmaydi")
}

// rateLimiter — IP bo'yicha fixed-window cheklov (brute-force himoyasi).
type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string]*winEntry
	limit  int
	window time.Duration
}

type winEntry struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{hits: make(map[string]*winEntry), limit: limit, window: window}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	w := rl.hits[ip]
	if w == nil || now.After(w.reset) {
		rl.hits[ip] = &winEntry{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.limit {
		return false
	}
	w.count++
	return true
}

func (rl *rateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		rl.mu.Lock()
		now := time.Now()
		for ip, w := range rl.hits {
			if now.After(w.reset) {
				delete(rl.hits, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r.RemoteAddr)
		if !rl.allow(ip) {
			writeErr(w, http.StatusTooManyRequests, "juda ko'p so'rov, biroz keyin urinib ko'ring")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP — RemoteAddr'dan port'ni olib tashlab, faqat IP'ni qaytaradi.
func clientIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
