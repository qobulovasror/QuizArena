-- name: GetRating :one
SELECT rating, games FROM user_rating WHERE user_id = $1 AND subject_id = $2;

-- name: UpsertRating :exec
INSERT INTO user_rating (user_id, subject_id, rating, games)
VALUES ($1, $2, $3, 1)
ON CONFLICT (user_id, subject_id)
DO UPDATE SET rating = EXCLUDED.rating, games = user_rating.games + 1, updated_at = now();

-- name: ListRatingsByUser :many
SELECT ur.rating, ur.games, s.slug AS subject_slug, s.name AS subject_name
FROM user_rating ur
JOIN subjects s ON s.id = ur.subject_id
WHERE ur.user_id = $1
ORDER BY ur.rating DESC;
