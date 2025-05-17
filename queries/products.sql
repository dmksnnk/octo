-- name: Products :many
SELECT * FROM products
WHERE products.deleted_at IS NULL;


-- name: ProductsWithPrices :many
SELECT sqlc.embed(products), sqlc.embed(prices)
FROM products
JOIN prices ON products.id = prices.product_id
WHERE products.deleted_at IS NULL 
AND prices.deleted_at IS NULL;


-- name: Product :one
SELECT * FROM products
WHERE products.id = @id
AND products.deleted_at IS NULL;

-- name: ProductWithPrice :one
SELECT sqlc.embed(products), sqlc.embed(prices)
FROM products
JOIN prices ON products.id = prices.product_id
WHERE products.id = @id
AND products.deleted_at IS NULL
AND prices.deleted_at IS NULL;