import MediaCard from "@/components/media-card/MediaCard";
import Pagination from "@/components/pagination/Pagination";
import { getRecommendations } from "@/lib/api/media";
import type { MediaType } from "@/lib/api/types";
import styles from "./Recommendations.module.css";

interface RecommendationsProps {
  id: number;
  type: MediaType;
  page: number;
}

const Recommendations = async ({ id, type, page }: RecommendationsProps) => {
  const response = await getRecommendations(type, id, page);
  const medias = response.results;

  if (medias.length === 0) {
    return null;
  }

  return (
    <>
      <h3 style={{ alignSelf: "start" }}>{type === "movie" ? "Похожие фильмы" : "Похожие сериалы"}</h3>
      <div className={styles.recommendations}>
        {medias.map((media) => (
          <div key={media.id}>
            <MediaCard className={styles.recommendationsFilm} {...media} type={type} />
          </div>
        ))}
      </div>
      <Pagination
        currentPage={page}
        totalPages={response.totalPages}
        pathname={`/details/${type}/${id}`}
        searchParams={{}}
        pageParam="recommendationsPage"
      />
    </>
  );
};

export default Recommendations;
