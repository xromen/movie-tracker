import { Media } from "@/lib/api/types"
import MediaCard from "../media-card/MediaCard"
import styles from "./MediaCatalog.module.css"

interface MediaCatalogProps {
    medias: Media[]
}

const MediaCatalog = ({ medias }: MediaCatalogProps) => {
    return (
        <div className={styles.grid}>
            {medias.map((media) => (
                <MediaCard key={media.id} {...media} />
            ))}
        </div>
    )
}

export default MediaCatalog