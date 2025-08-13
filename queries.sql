-- queries.sql - SQL queries for sqlc code generation

-- get all migrations
SELECT filename FROM migrations ORDER BY filename;

-- record a migration
INSERT INTO migrations (filename) VALUES (?);

-- name: GetUserByGitHubUID :one
SELECT * FROM users WHERE github_uid = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (github_uid, login, avatar_url) VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET login = ?, avatar_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetUserRooms :many
SELECT r.* FROM rooms r
JOIN room_memberships rm ON r.id = rm.room_id
WHERE rm.user_id = ?
ORDER BY r.created_at DESC;