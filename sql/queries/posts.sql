-- name: CreatePost :one
INSERT INTO posts (id,created_at,updated_at,title,url,description,published_at,feed_id)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: GetPosts :many
SELECT * FROM posts
LEFT JOIN feeds
ON posts.feed_id = feeds.id
LEFT JOIN feed_follows
ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
AND (feeds.name = sqlc.narg('name') OR sqlc.narg('name') IS NULL)
ORDER BY 
    (CASE WHEN sqlc.arg('sort') = 'created_at' AND sqlc.arg('is_desc')::boolean
        THEN posts.created_at END) DESC NULLS LAST,
    (CASE WHEN sqlc.arg('sort') = 'created_at' AND NOT sqlc.arg('is_desc')::boolean
        THEN posts.created_at END) ASC NULLS LAST,
    (CASE WHEN sqlc.arg('sort') = 'published_at' AND sqlc.arg('is_desc')::boolean
        THEN posts.published_at END) DESC NULLS LAST,
    (CASE WHEN sqlc.arg('sort') = 'published_at' AND NOT sqlc.arg('is_desc')::boolean
        THEN posts.published_at END) ASC NULLS LAST
LIMIT $2 OFFSET $3;