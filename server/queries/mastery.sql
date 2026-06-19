-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetMastery :one
SELECT * FROM user_mastery WHERE user_id = $1 AND category_id = $2;

-- name: UpsertMastery :exec
INSERT INTO user_mastery (user_id, category_id, mastery, attempts, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (user_id, category_id)
DO UPDATE SET mastery = EXCLUDED.mastery, attempts = EXCLUDED.attempts, updated_at = now();

-- name: ListMasteryByUser :many
SELECT um.mastery, um.attempts,
       c.slug AS category_slug, c.name AS category_name,
       s.slug AS subject_slug, s.name AS subject_name
FROM user_mastery um
JOIN categories c ON c.id = um.category_id
JOIN subjects s ON s.id = c.subject_id
WHERE um.user_id = $1
ORDER BY s.name, c.name;
