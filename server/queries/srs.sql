-- name: GetDueCards :many
SELECT q.id, q.type, q.prompt, q.options, q.correct, q.explanation
FROM srs_cards s
JOIN questions q ON q.id = s.question_id
JOIN categories c ON c.id = q.category_id
WHERE s.user_id = $1 AND c.subject_id = $2 AND s.due_at <= now()
ORDER BY s.due_at
LIMIT $3;

-- name: GetNewQuestions :many
SELECT q.id, q.type, q.prompt, q.options, q.correct, q.explanation
FROM questions q
JOIN categories c ON c.id = q.category_id
WHERE c.subject_id = $1
  AND NOT EXISTS (SELECT 1 FROM srs_cards s WHERE s.user_id = $2 AND s.question_id = q.id)
LIMIT $3;

-- name: GetSrsCard :one
SELECT * FROM srs_cards WHERE user_id = $1 AND question_id = $2;

-- name: UpsertSrsCard :exec
INSERT INTO srs_cards (user_id, question_id, ease, interval_days, repetitions, due_at, last_reviewed)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (user_id, question_id)
DO UPDATE SET ease = EXCLUDED.ease, interval_days = EXCLUDED.interval_days,
  repetitions = EXCLUDED.repetitions, due_at = EXCLUDED.due_at, last_reviewed = now();

-- name: CountDueCards :one
SELECT count(*) FROM srs_cards WHERE user_id = $1 AND due_at <= now();
