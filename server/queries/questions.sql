-- name: ListQuestionsByCategory :many
SELECT * FROM questions WHERE category_id = $1;

-- name: RandomQuestionsByCategory :many
SELECT * FROM questions
WHERE category_id = $1
ORDER BY random()
LIMIT $2;

-- name: CreateQuestion :one
INSERT INTO questions (
    category_id, type, prompt, options, correct, accept, media_url, explanation, difficulty, meta
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CountQuestionsByCategory :one
SELECT count(*) FROM questions WHERE category_id = $1;

-- name: RandomQuestionsBySubject :many
SELECT q.* FROM questions q
JOIN categories c ON c.id = q.category_id
WHERE c.subject_id = $1
ORDER BY random()
LIMIT $2;
