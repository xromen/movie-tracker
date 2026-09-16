DROP TABLE tv_episodes;
DROP TABLE movie_releases;

ALTER TABLE medias
    DROP COLUMN next_schedule_sync_at,
    DROP COLUMN schedule_synced_at,
    DROP COLUMN schedule_sync_error;