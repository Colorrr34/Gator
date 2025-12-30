-- name: CreateFeed :one
INSERT INTO feeds (id,created_at,updated_at,name,url,user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name, feeds.url, users.name
FROM feeds LEFT JOIN users
ON feeds.user_id = users.id;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET (updated_at, last_fetched_at) = (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
WHERE feeds.ID = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST, updated_at ASC
FETCH FIRST 1 ROWS ONLY;
