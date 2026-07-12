import Link from "next/link";
import Pagination from "@/components/pagination/Pagination";
import { getMediasList } from "@/lib/api/media";
import type { MovieList } from "@/lib/api/types";
import styles from "./MovieCatalog.module.css";
import MediaCatalog from "@/components/media-catalog/MediaCatalog";

const MOVIE_FILTERS: { label: string; value: MovieList }[] = [
  { label: "🎬 Смотрят сейчас", value: "now-playing" },
  { label: "🔥 Популярное", value: "popular" },
  { label: "⭐ Лучшие", value: "top-rated" },
  { label: "🍿 Ожидаемые", value: "upcoming" },
];

const isMovieFilter = (value?: string): value is MovieList =>
  value === "now-playing" || value === "popular" || value === "top-rated" || value === "upcoming";

interface MovieCatalogProps {
  filter?: string;
  page?: number;
}

const MovieCatalog = async ({ filter, page = 1 }: MovieCatalogProps) => {
  const activeFilter = isMovieFilter(filter) ? filter : "now-playing";
  const response = await getMediasList("movie", activeFilter, page);

  return (
    <div className={styles.page}>
      <h1 className="visuallyHidden">Фильмы</h1>

      <div className={styles.filters}>
        {MOVIE_FILTERS.map((item) => (
          <Link
            key={item.value}
            className={`${styles.filterButton} ${activeFilter === item.value ? styles.active : ""}`}
            href={`/movie?filter=${item.value}`}
          >
            {item.label}
          </Link>
        ))}
      </div>

      {response.results.length === 0 && <p>Ничего не найдено</p>}

      <MediaCatalog medias={response.results} />

      <Pagination
        currentPage={page}
        totalPages={response.totalPages}
        pathname="/movie"
        searchParams={{
          filter: activeFilter,
        }}
      />
    </div>
  );
};

export default MovieCatalog;
