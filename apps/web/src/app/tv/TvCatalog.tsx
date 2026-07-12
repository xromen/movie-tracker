import Link from "next/link";
import MediaCatalog from "@/components/media-catalog/MediaCatalog";
import Pagination from "@/components/pagination/Pagination";
import { getMediasList } from "@/lib/api/media";
import type { TvList } from "@/lib/api/types";
import styles from "./TvCatalog.module.css";

const TV_FILTERS: { label: string; value: TvList }[] = [
  { label: "📺 Сегодня в эфире", value: "airing-today" },
  { label: "🔥 Популярное", value: "popular" },
  { label: "⭐ Лучшие", value: "top-rated" },
  { label: "🗓️ Сейчас идут", value: "on-the-air" },
];

const isTvFilter = (value?: string): value is TvList =>
  value === "airing-today" || value === "popular" || value === "top-rated" || value === "on-the-air";

interface TvCatalogProps {
  filter?: string;
  page?: number;
}

const TvCatalog = async ({ filter, page = 1 }: TvCatalogProps) => {
  const activeFilter = isTvFilter(filter) ? filter : "airing-today";
  const response = await getMediasList("tv", activeFilter, page);
  const tvShows = response.results;

  return (
    <div className={styles.page}>
      <h1 className="visuallyHidden">Сериалы</h1>

      <div className={styles.filters}>
        {TV_FILTERS.map((item) => (
          <Link
            key={item.value}
            className={`${styles.filterButton} ${activeFilter === item.value ? styles.active : ""}`}
            href={`/tv?filter=${item.value}`}
          >
            {item.label}
          </Link>
        ))}
      </div>

      {tvShows.length === 0 && <p>Ничего не найдено</p>}

      <MediaCatalog medias={tvShows} />

      <Pagination
        currentPage={page}
        totalPages={response.totalPages}
        pathname="/tv"
        searchParams={{
          filter: activeFilter,
        }}
      />
    </div>
  );
};

export default TvCatalog;
