-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens
WHERE token = $1
AND expires_at > CURRENT_TIMESTAMP
AND revoked_at IS NULL;

-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE token = $1;
