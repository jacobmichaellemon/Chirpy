-- name: GetUserFromRefreshToken :one
SELECT user_id
FROM refreshtokens
WHERE token = $1
AND revoked_at IS NULL
AND expires_at > NOW();