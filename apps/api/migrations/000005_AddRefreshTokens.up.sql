ALTER TABLE users
    ADD COLUMN auth_version BIGINT NOT NULL DEFAULT 1;

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    replaced_by UUID REFERENCES refresh_tokens(id)
);

CREATE INDEX refresh_tokens_user_id_idx
    ON refresh_tokens(user_id);

CREATE INDEX refresh_tokens_family_id_idx
    ON refresh_tokens(family_id);

CREATE INDEX refresh_tokens_expires_at_idx
    ON refresh_tokens(expires_at);