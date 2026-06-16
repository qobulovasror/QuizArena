-- name: ListSubjects :many
SELECT * FROM subjects ORDER BY name;

-- name: GetSubjectBySlug :one
SELECT * FROM subjects WHERE slug = $1;

-- name: CreateSubject :one
INSERT INTO subjects (slug, name, icon)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListCategoriesBySubject :many
SELECT * FROM categories WHERE subject_id = $1 ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (subject_id, slug, name)
VALUES ($1, $2, $3)
RETURNING *;
