-- name: Avalability :one
SELECT * FROM availabilities
WHERE product_id = @product_id 
AND local_date = @local_date
AND deleted_at IS NULL;


-- name: AvalabilityWithPrice :one
SELECT sqlc.embed(availabilities), sqlc.embed(prices)
FROM availabilities
JOIN prices ON availabilities.product_id = prices.product_id
WHERE availabilities.product_id = @product_id
AND availabilities.local_date = @local_date
AND availabilities.deleted_at IS NULL
AND prices.deleted_at IS NULL;


-- name: AvalabilityRange :many
SELECT * FROM availabilities
WHERE product_id = @product_id 
AND local_date >= @local_date_start
AND local_date <= @local_date_end
AND deleted_at IS NULL;


-- name: AvalabilityWithPriceRange :many
SELECT sqlc.embed(availabilities), sqlc.embed(prices)
FROM availabilities
JOIN prices ON availabilities.product_id = prices.product_id
WHERE availabilities.product_id = @product_id 
AND availabilities.local_date >= @local_date_start
AND availabilities.local_date <= @local_date_end
AND availabilities.deleted_at IS NULL
AND prices.deleted_at IS NULL;