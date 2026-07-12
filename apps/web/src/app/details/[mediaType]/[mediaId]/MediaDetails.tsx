import { getMediaById } from "@/lib/api/media";
import { MediaType } from "@/lib/api/types";
import MediaHeader from "./MediaHeader";
import styles from "./MediaDetails.module.css"
import MediaCollectionCarousel from "./MediaCollectionCarousel";
import MediaRecommendationsCarousel from "./MediaRecommendationsCarousel";
import MediaSeasons from "./MediaSeasons";
import MediaVideosCarousel from "./MediaVideosCarousel";
import { getSession } from "@/lib/auth/session";

interface MediaDetailsProps {
    mediaId: number;
    mediaType: MediaType;
}

const MediaDetails = async ({ mediaId, mediaType }: MediaDetailsProps) => {
    const [media, session] = await Promise.all([
        getMediaById(mediaType, mediaId),
        getSession(),
    ])

    return (
        <div className={styles.wrapper}>
            <MediaHeader {...media} />
            <MediaVideosCarousel videos={media.videos} className={styles.carousel} />

            {mediaType === "movie" && media.collectionId && (
                <MediaCollectionCarousel collectionId={media.collectionId} className={styles.carousel} />
            )}

            {mediaType === "tv" && <MediaSeasons tvId={media.id} isAuthenticated={session.isAuthenticated} className={styles.carousel} />}
            <MediaRecommendationsCarousel mediaId={media.id} mediaType={mediaType} className={styles.carousel} />
        </div>
    )
}

export default MediaDetails;
