-- name: GetUserFromRefreshToken :one
SELECT users.*
FROM users
INNER JOIN refreshtokens ON users.id = refreshtokens.user_id
WHERE refreshtokens.token = $1
  AND refreshtokens.revoked_at IS NULL
  AND refreshtokens.expires_at > NOW();