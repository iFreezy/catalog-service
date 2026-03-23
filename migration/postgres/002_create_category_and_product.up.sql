CREATE TABLE IF NOT EXISTS category (
    id         BIGSERIAL,
    guid       UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS product (
    id            BIGSERIAL,
    guid          UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT,
    price         DECIMAL NOT NULL,
    category_guid UUID NOT NULL REFERENCES category(guid),
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL
);
