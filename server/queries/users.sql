-- name: CreateGuest :one
INSERT INTO users (is_guest) VALUES (true)
RETURNING *;

-- name: CreateAccount :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = $1;

-- name: CreateTelegramUser :one
INSERT INTO users (telegram_id) VALUES ($1)
RETURNING *;
