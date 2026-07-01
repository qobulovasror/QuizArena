# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

QuizArena — a multi-subject, multiplayer **real-time quiz platform**. One game engine serves three "intents": 🏆 competition (live multiplayer), 📚 learning (SRS / spaced repetition), 📊 assessment (mastery scoring). The full technical plan lives in `PLAN.md` (Uzbek), §-numbered; code comments reference those sections.

Note: the codebase, comments, and commit messages are written in **Uzbek**. Match that language in comments and commit messages.

## Repo layout

- `server/` — Go backend (`cmd/` entrypoints + `internal/*`)
- `client/` — TS monorepo (pnpm workspace): `apps/{web,telegram,native}` + `packages/{core,ui-web}`. Only `apps/web` is built out; the rest are placeholders.
- `shared/` — cross-language contracts: `protocol/` (WebSocket) and `openapi/` (REST). **`shared/protocol/README.md` is the source of truth** for the WS protocol; `server/internal/ws/protocol.go` and `shared/protocol/messages.ts` are kept in sync with it by hand.

## Commands

All `make` targets run from repo root (they `cd server` internally).

```bash
docker compose up -d postgres   # Postgres only (Redis is intentionally NOT used — see PLAN.md §1.1)
make migrate-up                 # goose migrations (needs DATABASE_URL env exported)
make seed                       # seed subjects/categories/questions (needs DATABASE_URL)
make run                        # run Go server locally
make build                      # build server -> server/bin/server
make sqlc                       # regenerate store/*.sql.go from queries/*.sql (needs sqlc)
make tidy                       # go mod tidy
make admin EMAIL=x@y.com        # promote a user to role=admin
make tools                      # install sqlc + goose CLIs
```

Tests (no test runner wrapper — use `go test` directly):
```bash
cd server && go test ./...                          # all
cd server && go test ./internal/game/...            # one package
cd server && go test ./internal/game -run TestName  # single test
```

Web frontend:
```bash
cd client/apps/web && npm install && npm run dev    # Vite dev server on :5173
npm run build                                       # tsc + vite build
```

### Local-dev port gotcha
The Vite dev proxy (`client/apps/web/vite.config.ts`) forwards `/api` and `/ws` to **`localhost:8099`**, but the server's default `PORT` is `8080`. For the web client to reach the backend in dev, run the server with `PORT=8099 make run` (or change the proxy target).

## Architecture

### Server-authoritative game engine
- `internal/game/engine.go` — the engine drives room lifecycle (lobby → countdown → per-question show/reveal → game over), scoring, and answer validation. **The client never sees correct answers**; options are sent with opaque IDs and the server validates.
- Transport vs. state vs. persistence are deliberately separate layers:
  - `internal/ws/` — WebSocket transport. `Hub` manages connections; `Client` is a connection; `Router` (`game/router.go`) maps incoming `Envelope` messages to engine methods. A `ws.Router` interface decouples transport from game logic.
  - `internal/state/` — **ephemeral** live game state, currently `MemStore` (in-memory). The `state.Store` interface is the seam: scaling swaps in a Redis impl without touching the engine (PLAN.md §2). Do not put durable data here.
  - `internal/persist/` — writes finished-game results to Postgres via the `Persister` interface (nil = don't persist; used in tests).
- `internal/game/registry.go` — `Registry` selects a **question Provider** per subject slug, with a fallback. Providers live in `internal/game/providers/`:
  - Generative/static-file providers: `english.go` (irregular verbs from `data/irregularVerb.json`), `math.go`, `sample.go` (fallback).
  - `general.go` — reads questions from the Postgres `questions` bank (populated by seed/admin). This is the path admin-managed questions flow through.
  - A `Provider` only implements `Questions(count int) ([]state.Question, error)`. Wiring happens in `cmd/server/main.go`.
- `internal/game/qtype/` — question-type strategies. Each type (`mcq`, `trueFalse`, `numeric`, `typeAnswer`) differs **only in `Validate`**; render/reveal are generic. Adding a type = one strategy struct + a case in `For()`.

### Persistence (Postgres + sqlc)
- Schema = goose migrations in `server/migrations/` (numbered). Queries = `server/queries/*.sql`. `make sqlc` generates typed Go into `internal/store/*.sql.go` per `server/sqlc.yaml`.
- **Workflow for DB changes:** edit/add a migration, write the query in `queries/`, run `make sqlc`, then use the generated `store.Queries` method. Don't hand-edit `internal/store/*.sql.go`.
- sqlc overrides map `uuid` → `google/uuid.UUID` and `timestamptz` → `time.Time` (nullable → pointers).

### HTTP API & auth
- `internal/httpapi/router.go` is the single place all routes are registered (chi router). Handlers are gated by dependency presence (`d.Auth != nil`, `d.Queries != nil`) and by middleware: `requireAuth`, `requireAdmin` (RBAC role=admin), and a per-IP rate limiter on `/api/auth`.
- `internal/auth/` — guest + account + Telegram login, issuing JWTs (`token.go`). The same JWT authenticates the WS connection via `?token=` query or `Authorization: Bearer` (see `wsAuth` in router.go); guests connect tokenless and supply `displayName` on join.
- Route groups: `/api/auth/*`, `/api/subjects/*`, authed `/api/me/*` (history, SRS due/review, assessment, mastery), and `/api/admin/*` (subject/category/question CRUD).

### Learning & assessment columns
- `internal/learn/srs.go` — SM-2 spaced-repetition scheduling (📚). `internal/assess/mastery.go` — EMA-based mastery scoring (📊). Both are pure logic with unit tests, surfaced through `/api/me/*` handlers.

### Client (web)
- `apps/web/src/core/` is the shared logic layer: `store.ts` (Zustand store holding the WS socket, auth, and game state with auto-reconnect), `protocol.ts` (TS mirror of the WS contract), `api.ts` (REST client), `i18n.ts` (uz default + en), `telegram.ts` (Mini App auto-login).
- `App.tsx` is a screen router driven by store state (auth → home → lobby → play → result), not a URL router. Pages live in `src/pages/`.
- UI: Tailwind + shadcn-style components in `src/components/ui/`.

## Conventions
- Adding a subject: register a `Provider` in `cmd/server/main.go` against its slug.
- Adding a WS message: update `shared/protocol/README.md` first, then mirror in `protocol.go` (Go) and `messages.ts` (TS), then handle in `game/router.go` + engine and in the client store.
- Redis is intentionally absent until the scaling phase — keep live state behind the `state.Store` interface.
