CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    username      VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TYPE media_type AS ENUM ('tv', 'movie');

CREATE TABLE medias
(
    id           BIGSERIAL PRIMARY KEY,
    tmdb_id      INTEGER      NOT NULL UNIQUE,
    title        VARCHAR(500) NOT NULL,
    overview     TEXT,
    poster_path  VARCHAR(500),
    release_date DATE,
    vote_average NUMERIC(10, 2),
    media_type   media_type NOT NULL,
    metadata     JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TYPE watch_status AS ENUM ('watched', 'want_to_watch', 'favorite');

CREATE TABLE user_medias
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users (id),
    media_id   BIGINT       NOT NULL REFERENCES medias (tmdb_id),
    status     watch_status NOT NULL,
    rating     SMALLINT CHECK (rating >= 1 AND rating <= 10),
    watched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, media_id)
);

CREATE INDEX idx_user_medias_user_id ON user_medias (user_id);
CREATE INDEX idx_user_medias_status ON user_medias (user_id, status);
CREATE INDEX idx_medias_tmdb_id ON medias (tmdb_id);