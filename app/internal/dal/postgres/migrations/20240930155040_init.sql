-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE "order" (
    id           UUID           NOT NULL, -- UUID primary key.
    user_id      INT            NOT NULL, -- User ID.
    number_order SERIAL         NOT NULL, -- Number Order (serial).
    status       TEXT           NOT NULL, -- status order(create, accepted, sent, delivered).
    type_product TEXT           NOT NULL, -- 2 types(breakable , unbreakable).
    price        NUMERIC(64, 8) NOT NULL, -- Price.
    item         INT            NOT NULL, -- Item amount.
    packs        JSONB          NOT NULL DEFAULT '[]', -- Packs with size and count (JSON).
    created_at   TIMESTAMPTZ    NOT NULL, -- Date created order.
    updated_at   TIMESTAMPTZ    NOT NULL, -- Date updated order.
    CONSTRAINT order_id_pk PRIMARY KEY (id)
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
