"use client";

import { useEffect, useRef, useState } from "react";
import type { Video } from "@/lib/api/types";
import YoutubePlayer from "./YoutubePlayer";
import styles from "./TrailerModal.module.css";

interface TrailerButtonProps {
  videos: Video[];
  className?: string;
}

const TrailerButton = ({ videos, className }: TrailerButtonProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const thumbnailsRef = useRef<HTMLDivElement>(null);
  const thumbnailRefs = useRef<(HTMLButtonElement | null)[]>([]);

  useEffect(() => {
    if (!isOpen) return;

    const handleKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setIsOpen(false);
      if (event.key === "ArrowLeft") setActiveIndex((index) => Math.max(0, index - 1));
      if (event.key === "ArrowRight") setActiveIndex((index) => Math.min(videos.length - 1, index + 1));
    };

    document.addEventListener("keydown", handleKey);
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKey);
      document.body.style.overflow = "";
    };
  }, [isOpen, videos.length]);

  useEffect(() => {
    const element = thumbnailRefs.current[activeIndex];
    element?.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" });
  }, [activeIndex]);

  useEffect(() => {
    const element = thumbnailsRef.current;
    if (!element || !isOpen) return;

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();
      element.scrollLeft += event.deltaY;
    };

    element.addEventListener("wheel", handleWheel, { passive: false });

    return () => element.removeEventListener("wheel", handleWheel);
  }, [isOpen]);

  if (videos.length === 0) {
    return null;
  }

  const activeVideo = videos[activeIndex];

  return (
    <>
      <button type="button" className={className} onClick={() => setIsOpen(true)}>
        Смотреть трейлер
      </button>

      {isOpen && (
        <div className={styles.overlay} onClick={() => setIsOpen(false)}>
          <div className={styles.modal} onClick={(event) => event.stopPropagation()}>
            <div className={styles.header}>
              <h3 className={styles.title}>Трейлеры</h3>
              <button className={styles.closeButton} onClick={() => setIsOpen(false)}>
                x
              </button>
            </div>

            <div className={styles.player}>
              <YoutubePlayer key={activeVideo.key} videoKey={activeVideo.key} title={activeVideo.name} />
            </div>

            {videos.length > 1 && (
              <div className={styles.carousel}>
                <button className={styles.arrowButton} onClick={() => setActiveIndex((index) => Math.max(0, index - 1))} disabled={activeIndex === 0}>
                  ‹
                </button>

                <div className={styles.thumbnails} ref={thumbnailsRef}>
                  {videos.map((video, index) => (
                    <button
                      key={video.id}
                      className={`${styles.thumbnail} ${index === activeIndex ? styles.thumbnailActive : ""}`}
                      ref={(element) => {
                        thumbnailRefs.current[index] = element;
                      }}
                      onClick={() => setActiveIndex(index)}
                    >
                      <img src={`https://img.youtube.com/vi/${video.key}/mqdefault.jpg`} alt={video.name} className={styles.thumbnailImage} />
                      <span className={styles.thumbnailIndex}>{index + 1}</span>
                    </button>
                  ))}
                </div>

                <button
                  className={styles.arrowButton}
                  onClick={() => setActiveIndex((index) => Math.min(videos.length - 1, index + 1))}
                  disabled={activeIndex === videos.length - 1}
                >
                  ›
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
};

export default TrailerButton;
