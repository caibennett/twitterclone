-- name: ListPostsAndUsers :many
SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.name, users.username FROM posts
INNER JOIN users ON posts.user_id=users.id
ORDER BY posts.created_at DESC
LIMIT 100;

-- name: SearchPosts :many
SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.name, users.username FROM posts
INNER JOIN users ON posts.user_id=users.id
WHERE posts.content LIKE concat("%", ?, "%")
LIMIT 100;

-- name: CreatePost :exec
INSERT INTO posts (
    id, user_id, content, created_at, updated_at
) VALUES (
  ?, ?, ?, unixepoch('now'), unixepoch('now')
);