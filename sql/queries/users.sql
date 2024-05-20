-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUsername :one
SELECT username FROM users
WHERE username = ? LIMIT 1;

-- name: GetFromUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (
  id, username, password, created_at, updated_at
) VALUES (
  ?, ?, ?, unixepoch('now'), unixepoch('now')
);

-- name: SetName :exec
UPDATE users
SET name = ?
WHERE id = ?;