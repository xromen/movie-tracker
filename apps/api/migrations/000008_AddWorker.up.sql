ALTER TABLE medias
    ADD COLUMN next_schedule_sync_at TIMESTAMPTZ,
    ADD COLUMN schedule_synced_at TIMESTAMPTZ,
    ADD COLUMN schedule_sync_error TEXT;

CREATE INDEX idx_medias_next_schedule_sync
    ON medias (next_schedule_sync_at)
    WHERE next_schedule_sync_at IS NOT NULL;

CREATE TABLE movie_releases
(
    media_id     BIGINT      NOT NULL REFERENCES medias (id),
    region       CHAR(2)     NOT NULL,
    release_type SMALLINT    NOT NULL CHECK (release_type BETWEEN 1 AND 6),
    release_at   TIMESTAMPTZ NOT NULL,
    certification VARCHAR(32),
    PRIMARY KEY (media_id, region, release_type, release_at)
);

CREATE TABLE tv_episodes
(
    tmdb_episode_id BIGINT PRIMARY KEY,
    media_id        BIGINT       NOT NULL REFERENCES medias (id),
    season_number   INTEGER      NOT NULL,
    episode_number  INTEGER      NOT NULL,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    overview        TEXT         NOT NULL DEFAULT '',
    air_date        DATE,
    still_path      VARCHAR(500),
    runtime         INTEGER,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (media_id, season_number, episode_number)
);

CREATE INDEX idx_movie_releases_date
    ON movie_releases (release_at);

CREATE INDEX idx_tv_episodes_air_date
    ON tv_episodes (air_date);