// Command server — QuizArena backend kirish nuqtasi.
//
// Bosqich 1 (qisman): config + httpapi (health + /ws) + ws.Hub.
// Hali yo'q: store (pgx/sqlc), auth, game engine — keyingi qadamlar.
// Qarang: PLAN.md §3 (arxitektura), §10 (struktura).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azizbek12234/quizarena/server/internal/auth"
	"github.com/azizbek12234/quizarena/server/internal/config"
	"github.com/azizbek12234/quizarena/server/internal/game"
	"github.com/azizbek12234/quizarena/server/internal/game/providers"
	"github.com/azizbek12234/quizarena/server/internal/httpapi"
	"github.com/azizbek12234/quizarena/server/internal/persist"
	"github.com/azizbek12234/quizarena/server/internal/state"
	"github.com/azizbek12234/quizarena/server/internal/store"
	"github.com/azizbek12234/quizarena/server/internal/ws"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	// Postgres (doimiy ma'lumot — auth uchun zarur).
	pool, err := store.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("postgres ulanmadi (docker compose up postgres ishga tushiringmi?)", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	queries := store.New(pool)

	// Auth (mehmon + akkaunt + JWT).
	tokens := auth.NewTokenManager(cfg.JWTSecret, 7*24*time.Hour)
	authSvc := auth.NewService(queries, tokens)

	// Jonli o'yin (in-memory state + engine).
	hub := ws.NewHub(logger)
	liveStore := state.NewMemStore()

	registry := game.NewRegistry(providers.NewSample()) // noma'lum soha → Sample
	if ev, err := providers.NewEnglishVerb(); err != nil {
		logger.Error("english provider yuklanmadi", "err", err)
	} else {
		registry.Register("english", ev)
	}
	registry.Register("math", providers.NewMath())
	if gen, err := queries.GetSubjectBySlug(context.Background(), "general"); err != nil {
		logger.Warn("general soha topilmadi — `make seed` ishga tushiring", "err", err)
	} else {
		registry.Register("general", providers.NewGeneral(queries, gen.ID))
	}

	persister := persist.NewDB(queries)
	engine := game.NewEngine(hub, liveStore, registry, persister, logger)
	gameRouter := game.NewRouter(engine)

	handler := httpapi.Router(httpapi.Deps{
		Cfg: cfg, Hub: hub, WSRouter: gameRouter, Auth: authSvc, Queries: queries, Logger: logger,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Signal'da to'xtash konteksti.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("QuizArena server tinglayapti", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server xatosi", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("to'xtatilmoqda...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown xatosi", "err", err)
	}
	logger.Info("to'xtatildi")
}
