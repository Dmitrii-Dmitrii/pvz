CREATE TYPE city AS enum (
    'moscow',
    'spb',
    'kazan'
    );

CREATE TYPE reception_status AS enum (
    'in_progress',
    'close'
    );

CREATE TYPE product_type AS enum (
    'electronics',
    'clothes',
    'shoes'
    );

CREATE TABLE IF NOT EXISTS pvzs
(
    id            UUID PRIMARY KEY,
    register_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    city          city      NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions
(
    id             UUID PRIMARY KEY,
    reception_time TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    pvz_id         UUID            NOT NULL,
    status          reception_status NOT NULL,
    FOREIGN KEY (pvz_id) REFERENCES pvzs (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS products
(
    id           UUID PRIMARY KEY,
    adding_time  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    product_type product_type NOT NULL,
    reception_id UUID         NOT NULL,
    FOREIGN KEY (reception_id) REFERENCES receptions (id) ON DELETE CASCADE
);

CREATE INDEX idx_receptions_pvz_id ON receptions (pvz_id);
CREATE INDEX idx_products_reception_id ON products (reception_id);