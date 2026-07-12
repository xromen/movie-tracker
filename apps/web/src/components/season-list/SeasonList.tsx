"use client";

import { useState, useTransition } from "react";
import { getSeasonEpisodes, setEpisodeWatched, setSeasonWatched as saveSeasonWatched } from "@/lib/api/media";
import type { Episode, Season } from "@/lib/api/types";
import styles from "./SeasonList.module.css";

interface SeasonListProps {
  tvId: number;
  seasons: Season[];
  isAuthenticated: boolean;
}

const getColor = (value?: number): string => {
  if (!value) return "";
  if (value >= 7) return "#1D9E75";
  if (value >= 5) return "#EF9F27";
  return "#E24B4A";
};

const getTimeFromMinutes = (totalMinutes: number): string => {
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;

  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:00`;
};

const SeasonList = ({ tvId, seasons, isAuthenticated }: SeasonListProps) => {
  if (seasons.length === 0) return null;

  return (
    <section className={styles.section} aria-labelledby="seasons-title">
      <h2 id="seasons-title">Сезоны</h2>

      <div className={styles.list}>
        {seasons.map((season) => (
          <SeasonItem key={season.id} tvId={tvId} season={season} isAuthenticated={isAuthenticated} />
        ))}
      </div>
    </section>
  );
};

const SeasonItem = ({ tvId, season, isAuthenticated }: { tvId: number; season: Season; isAuthenticated: boolean }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [episodes, setEpisodes] = useState<Episode[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [seasonWatched, setSeasonWatched] = useState(season.isWatched);
  const [isError, setIsError] = useState(false);
  const [isPending, startTransition] = useTransition();

  const loadEpisodes = (nextPage: number) => {
    startTransition(async () => {
      setIsError(false);

      try {
        const data = await getSeasonEpisodes(tvId, season.seasonNumber, nextPage);

        setEpisodes(data.episodes);
        setTotalPages(data.totalPages);
        setPage(nextPage);
      } catch {
        setIsError(true);
      }
    });
  };

  const toggleOpen = () => {
    const nextOpen = !isOpen;
    setIsOpen(nextOpen);

    if (nextOpen && episodes.length === 0) {
      loadEpisodes(1);
    }
  };

  const toggleSeasonWatched = async () => {
    const nextWatched = !seasonWatched;

    await saveSeasonWatched(tvId, season.seasonNumber, nextWatched);
    setSeasonWatched(nextWatched);
  };

  const toggleEpisodeWatched = async (episode: Episode) => {
    await setEpisodeWatched(tvId, season.seasonNumber, episode.episodeNumber, !episode.isWatched);
    loadEpisodes(page);
  };

  return (
    <article className={styles.season}>
      <div className={styles.seasonButton} role="button" onClick={toggleOpen} aria-expanded={isOpen}>
        <div className={styles.poster}>
          {season.posterPath ? <img src={season.posterPath} alt={`Постер: ${season.title}`} /> : <div className={styles.posterPlaceholder}>🎬</div>}
        </div>

        <div className={styles.seasonInfo}>
          <h3 className={styles.seasonTitle}>{season.title}</h3>
          <div className={styles.meta}>
            {season.voteAverage !== 0 && (
              <span className={`${styles.metaItem} ${styles.rating}`} style={{ background: getColor(season.voteAverage) }}>
                {season.voteAverage.toFixed(1)}
              </span>
            )}
            {season.releaseDate && <span className={styles.metaItem}>{season.releaseDate.toLocaleDateString("ru")}</span>}
            <span className={styles.metaItem}>{season.episodeCount} эпизодов</span>
          </div>
          {season.overview && <span className={styles.seasonOverview}>{season.overview}</span>}
        </div>

        {isAuthenticated && seasonWatched !== undefined && (
          <button
            type="button"
            className={`${styles.watchedToggle} ${seasonWatched ? styles.watchedToggleActive : ""}`}
            onClick={(event) => {
              event.stopPropagation();
              void toggleSeasonWatched();
            }}
            aria-label="Переключить просмотр сезона"
          >
            {seasonWatched ? "✓" : ""}
          </button>
        )}
      </div>

      <div className={`${styles.episodesWrapper} ${isOpen ? styles.episodesWrapperOpen : ""}`}>
        <div className={styles.episodes}>
          {isPending && <p className={styles.message}>Загрузка серий...</p>}
          {isError && <p className={styles.error}>Не удалось загрузить список серий.</p>}

          {episodes.map((episode) => (
            <div className={styles.episode} key={episode.id}>
              <div className={styles.episodeImage}>
                {episode.posterPath ? <img src={episode.posterPath} alt={`Кадр из серии: ${episode.title}`} /> : <div className={styles.episodePlaceholder}>🎬</div>}
                <div className={styles.runtime}>{getTimeFromMinutes(episode.runtime)}</div>

                {isAuthenticated && episode.isWatched !== undefined && (
                  <button
                    type="button"
                    className={`${styles.watchedToggle} ${styles.watchedToggleEpisode} ${episode.isWatched ? styles.watchedToggleActive : ""}`}
                    onClick={() => void toggleEpisodeWatched(episode)}
                    aria-label="Переключить просмотр серии"
                  >
                    {episode.isWatched ? "✓" : ""}
                  </button>
                )}
              </div>

              <div className={styles.episodeInfo}>
                <h4 className={styles.seasonTitle}>
                  {episode.episodeNumber}. {episode.title}
                </h4>
                <div className={styles.meta}>
                  {episode.voteAverage !== 0 && (
                    <span className={`${styles.metaItem} ${styles.rating}`} style={{ background: getColor(episode.voteAverage) }}>
                      {episode.voteAverage.toFixed(1)}
                    </span>
                  )}
                  {episode.releaseDate && <span className={styles.metaItem}>{episode.releaseDate.toLocaleDateString("ru")}</span>}
                </div>
                {episode.overview && <p>{episode.overview}</p>}
              </div>
            </div>
          ))}

          {!isPending && !isError && isOpen && episodes.length === 0 && <p className={styles.message}>Серии пока не найдены.</p>}

          {totalPages > 1 && (
            <div className={styles.paginationActions}>
              <button className={styles.loadMoreButton} onClick={() => loadEpisodes(page - 1)} disabled={page <= 1 || isPending}>
                Назад
              </button>
              <span className={styles.message}>
                {page} / {totalPages}
              </span>
              <button className={styles.loadMoreButton} onClick={() => loadEpisodes(page + 1)} disabled={page >= totalPages || isPending}>
                Вперед
              </button>
            </div>
          )}
        </div>
      </div>
    </article>
  );
};

export default SeasonList;
