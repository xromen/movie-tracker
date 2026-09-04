ALTER TABLE users
    ADD COLUMN telegram_id bigint UNIQUE;

CREATE TABLE telegram_binding_tokens
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users (id),
    token      varchar(10) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);