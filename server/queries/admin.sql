-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- name: SetRoleByEmail :exec
UPDATE users SET role = $2 WHERE email = $1;
