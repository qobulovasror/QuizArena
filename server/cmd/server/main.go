// Command server — QuizArena backend kirish nuqtasi.
//
// Bosqich 0: skelet. Hozircha biznes-logika yo'q. Keyingi bosqichlarda shu yerda
// bootstrap ketma-ketligi ulanadi: config → store (pgx/sqlc) → ws.Hub → httpapi.Router.
// Qarang: PLAN.md §3 (arxitektura), §10 (struktura).
package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("QuizArena server — skelet (Bosqich 0)",
		"port", port,
		"todo", "config, store(pgx/sqlc), ws.Hub, httpapi.Router",
	)

	// TODO(Bosqich 1):
	//   cfg := config.Load()
	//   db := store.Connect(cfg.DatabaseURL)
	//   hub := ws.NewHub()
	//   r := httpapi.Router(cfg, db, hub)
	//   http.ListenAndServe(":"+port, r)
}
