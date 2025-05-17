-- name: UserByAPIKey :one
SELECT * FROM users
WHERE api_key = sqlc.arg('api_key')::VARCHAR
AND deleted_at IS NULL;