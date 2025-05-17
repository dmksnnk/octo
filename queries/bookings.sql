-- name: CreateBooking :one
WITH reservation AS (
    UPDATE availabilities
    SET vacancies = vacancies - @units, 
        updated_at = CURRENT_TIMESTAMP
    WHERE id = sqlc.arg('availability_id')::INTEGER
    AND product_id = sqlc.arg('product_id')::INTEGER
    AND deleted_at IS NULL
    AND vacancies >= @units
    RETURNING product_id, id AS availability_id
),
reserved_booking AS (
    INSERT INTO bookings (product_id, availability_id, user_id, status)
    SELECT reservation.product_id, reservation.availability_id, @user_id, 'RESERVED' 
    FROM reservation
    RETURNING id
),
new_units AS (
    INSERT INTO units (booking_id)
    SELECT reserved_booking.id 
    FROM reserved_booking, generate_series(1, @units) -- creating units number of rows
)
SELECT *
FROM reserved_booking
;


-- name: BookingForUpdate :many
SELECT bookings.status, units.id as unit_id
FROM bookings
JOIN units ON units.booking_id = bookings.id
WHERE bookings.id = @id
AND bookings.user_id = @user_id
AND bookings.deleted_at IS NULL
AND units.deleted_at IS NULL
FOR UPDATE;

-- name: ConfirmBooking :exec
UPDATE bookings
SET status = 'CONFIRMED',
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: SetUnitTicket :exec
UPDATE units
SET ticket = @ticket,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: Booking :many
SELECT sqlc.embed(bookings), sqlc.embed(units)
FROM bookings
LEFT JOIN units ON units.booking_id = bookings.id
WHERE bookings.id = @id
AND bookings.user_id = @user_id
AND bookings.deleted_at IS NULL
AND units.deleted_at IS NULL;

-- name: BookingWithPrice :many
SELECT sqlc.embed(bookings), sqlc.embed(units), sqlc.embed(prices)
FROM bookings
LEFT JOIN units ON units.booking_id = bookings.id
JOIN prices ON  prices.product_id = bookings.product_id
WHERE bookings.id = @id
AND bookings.user_id = @user_id
AND bookings.deleted_at IS NULL
AND units.deleted_at IS NULL
AND prices.deleted_at IS NULL;
