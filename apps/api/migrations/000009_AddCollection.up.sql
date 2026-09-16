CREATE TABLE movie_collections
(
    tmdb_id               BIGINT PRIMARY KEY,
    name                  VARCHAR(500),
    next_schedule_sync_at TIMESTAMPTZ,
    schedule_synced_at    TIMESTAMPTZ,
    schedule_sync_error   TEXT
);

ALTER TABLE medias
    ADD COLUMN collection_tmdb_id BIGINT
        REFERENCES movie_collections (tmdb_id);

CREATE INDEX idx_medias_collection
    ON medias (collection_tmdb_id)
    WHERE collection_tmdb_id IS NOT NULL;