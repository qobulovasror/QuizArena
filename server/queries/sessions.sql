-- name: CreateSession :one
INSERT INTO game_sessions (
    code, host_user_id, subject_id, category_id, mode, opponent, question_count, time_per_q
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSessionByCode :one
SELECT * FROM game_sessions WHERE code = $1;

-- name: GetSessionByID :one
SELECT * FROM game_sessions WHERE id = $1;

-- name: SetSessionRunning :exec
UPDATE game_sessions SET status = 'running', started_at = now() WHERE id = $1;

-- name: SetSessionFinished :exec
UPDATE game_sessions SET status = 'finished', finished_at = now() WHERE id = $1;

-- name: UpsertResult :one
INSERT INTO game_results (session_id, user_id, score, correct_cnt, rank)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (session_id, user_id)
DO UPDATE SET score = EXCLUDED.score, correct_cnt = EXCLUDED.correct_cnt, rank = EXCLUDED.rank
RETURNING *;

-- name: ListResultsBySession :many
SELECT * FROM game_results WHERE session_id = $1 ORDER BY rank;

-- name: InsertAnswerLog :exec
INSERT INTO answers_log (session_id, user_id, question_id, given, is_correct, time_ms)
VALUES ($1, $2, $3, $4, $5, $6);
