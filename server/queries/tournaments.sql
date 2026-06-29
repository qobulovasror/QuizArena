-- name: CreateTournament :one
INSERT INTO tournaments (title, subject_id, question_count, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListTournaments :many
SELECT t.*, s.slug AS subject_slug, s.name AS subject_name
FROM tournaments t
JOIN subjects s ON s.id = t.subject_id
ORDER BY t.starts_at DESC;

-- name: GetTournament :one
SELECT * FROM tournaments WHERE id = $1;

-- name: UpsertEntry :exec
-- Eng yaxshi ball saqlanadi (GREATEST).
INSERT INTO tournament_entries (tournament_id, user_id, score)
VALUES ($1, $2, $3)
ON CONFLICT (tournament_id, user_id)
DO UPDATE SET score = GREATEST(tournament_entries.score, EXCLUDED.score), updated_at = now();

-- name: ListEntries :many
SELECT te.score, u.username, te.user_id
FROM tournament_entries te
JOIN users u ON u.id = te.user_id
WHERE te.tournament_id = $1
ORDER BY te.score DESC, te.updated_at ASC
LIMIT 50;
