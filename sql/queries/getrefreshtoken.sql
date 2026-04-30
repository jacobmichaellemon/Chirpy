-- name: LookupRefreshToken :one
SELECT * 
FROM refreshtokens
WHERE token = $1;