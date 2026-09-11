DROP INDEX idx_medias_collection;
ALTER TABLE medias
    DROP COLUMN collection_tmdb_id;
DROP TABLE movie_collections;