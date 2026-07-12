CREATE TABLE user_tv_episodes
(
    user_id         BIGINT      NOT NULL REFERENCES users (id),
    tv_show_tmdb_id INTEGER     NOT NULL CHECK (tv_show_tmdb_id > 0),
    season_number   INTEGER     NOT NULL CHECK (season_number >= 0),
    episode_number  INTEGER     NOT NULL CHECK (episode_number > 0),
    watched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, tv_show_tmdb_id, season_number, episode_number)
);