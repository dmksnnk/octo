-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    email VARCHAR NOT NULL UNIQUE,
    api_key VARCHAR
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    name VARCHAR NOT NULL,
    capacity INTEGER NOT NULL
        CONSTRAINT non_negative_capacity CHECK ( capacity >= 0 )
);

CREATE INDEX idx_active_products ON products (id, deleted_at)
    WHERE deleted_at IS NULL;


CREATE TABLE availabilities (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    product_id INTEGER NOT NULL REFERENCES products(id),

    local_date DATE NOT NULL,
    vacancies INTEGER NOT NULL
        CONSTRAINT non_negative_vacancies CHECK ( vacancies >= 0 )
);

CREATE INDEX idx_active_availabilities_by_product_local_date ON availabilities (product_id, local_date, deleted_at)
    WHERE deleted_at IS NULL;

CREATE TYPE booking_status AS ENUM (
    'RESERVED',
    'CONFIRMED'
);

CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    product_id INTEGER NOT NULL REFERENCES products(id), -- which product was booked
    availability_id INTEGER NOT NULL REFERENCES availabilities(id), -- which availability was booked
    user_id INTEGER NOT NULL REFERENCES users(id), -- who booked the product
    status booking_status NOT NULL
);

CREATE TABLE units (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    booking_id BIGINT NOT NULL REFERENCES bookings(id),
    ticket VARCHAR
);

CREATE TABLE prices (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    price INTEGER NOT NULL
        CONSTRAINT non_negative_price CHECK ( price >= 0 ),
    currency CHAR(3) NOT NULL,
    product_id INTEGER NOT NULL REFERENCES products(id)
);

CREATE INDEX idx_active_prices_by_product ON prices (product_id, deleted_at)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_active_prices_by_product;
DROP INDEX IF EXISTS idx_active_availabilities_by_product_local_date;
DROP INDEX IF EXISTS idx_active_products;

DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS availabilities;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS booking_status;
-- +goose StatementEnd
