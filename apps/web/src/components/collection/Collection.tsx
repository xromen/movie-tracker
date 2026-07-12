import MediaCard from "@/components/media-card/MediaCard";
import { getCollection } from "@/lib/api/collection";
import styles from "./Collection.module.css";

interface CollectionProps {
  id: number;
}

const Collection = async ({ id }: CollectionProps) => {
  const collection = await getCollection(id);
  const parts = collection.parts
    .slice()
    .sort((a, b) => {
      if (!a.releaseDate && !b.releaseDate) return 0;
      if (!a.releaseDate) return 1;
      if (!b.releaseDate) return -1;
      return a.releaseDate.getTime() - b.releaseDate.getTime();
    });

  return (
    <>
      <h3 style={{ alignSelf: "start" }}>{collection.name}</h3>
      <div className={styles.parts}>
        {parts.map((media) => (
          <div key={`${media.type}-${media.id}`}>
            <MediaCard className={styles.part} {...media} type={media.type} />
          </div>
        ))}
      </div>
    </>
  );
};

export default Collection;
