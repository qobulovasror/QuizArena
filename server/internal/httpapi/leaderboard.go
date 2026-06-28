package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

type leaderboardHandler struct {
	q      *store.Queries
	logger *slog.Logger
}

type leaderboardRow struct {
	Username   string  `json:"username"`
	TotalScore float64 `json:"totalScore"`
	Games      int32   `json:"games"`
	Correct    int32   `json:"correct"`
}

// global — global reyting (ochiq, auth shart emas): top 20 akkaunt umumiy ball bo'yicha.
func (h *leaderboardHandler) global(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.GlobalLeaderboard(r.Context())
	if err != nil {
		h.logger.Error("leaderboard", "err", err)
		writeErr(w, http.StatusInternalServerError, "reytingni olishda xato")
		return
	}
	items := make([]leaderboardRow, 0, len(rows))
	for _, row := range rows {
		name := ""
		if row.Username != nil {
			name = *row.Username
		}
		items = append(items, leaderboardRow{
			Username: name, TotalScore: row.TotalScore, Games: row.Games, Correct: row.Correct,
		})
	}
	writeJSON(w, http.StatusOK, items)
}
