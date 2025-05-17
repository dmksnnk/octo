-- name: InsertProduct :one
-- used in tests
INSERT INTO products (name, capacity, deleted_at) 
VALUES (@name, @capacity, @deleted_at) 
RETURNING *;


-- name: InsertPrice :one
-- used in tests
INSERT INTO prices (price, currency, product_id, deleted_at)
VALUES (@price, @currency, @product_id, @deleted_at) 
RETURNING *;

-- name: InsertAvailability :one
INSERT INTO availabilities (product_id, local_date, vacancies, deleted_at) 
VALUES (@product_id, @local_date, @vacancies, @deleted_at)
RETURNING *;

-- name: InsertUser :one
-- used in tests
INSERT INTO users (email, api_key)
VALUES (@email, @api_key)
RETURNING *;
