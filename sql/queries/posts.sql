-- name: ListPostsAndUsers :many
SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.name, users.username FROM posts
INNER JOIN users ON posts.user_id=users.id
LIMIT 100;