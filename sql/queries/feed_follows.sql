-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows(id,created_at,updated_at,user_id,feed_id)
    VALUES(
        $1,$2,$3,$4,$5
    )
    RETURNING *
)
SELECT inserted.*, users.name AS user_name, feeds.name AS feed_name
FROM inserted INNER JOIN feeds
ON inserted.feed_id = feeds.id
INNER JOIN users
ON feeds.user_id = users.id;
    
-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, users.name, feeds.name
FROM feed_follows LEFT JOIN feeds
ON feed_follows.feed_id = feeds.id
LEFT JOIN users
ON feeds.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1
AND feed_follows.feed_id = $2;