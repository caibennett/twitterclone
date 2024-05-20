-- name: GetUserFromSession :one
SELECT users.*
FROM sessions
INNER JOIN users ON sessions.user_id = users.id
WHERE sessions.token = ? AND sessions.expire_at > unixepoch('now');

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE sessions.token = ?;

-- name: ListSessions :many
SELECT * FROM sessions;

-- name: CreateSession :exec
INSERT INTO sessions (
    token, user_id, ip_address, expire_at
) VALUES (
    ?, ?, ?, ?
);

-- name: DeleteExpired :exec
DELETE FROM sessions WHERE expire_at <= ?;