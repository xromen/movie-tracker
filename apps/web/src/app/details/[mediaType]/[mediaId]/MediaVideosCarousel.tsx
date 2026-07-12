"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { ChevronLeft, ChevronRight, Play, X } from "lucide-react"
import Carousel from "@/components/carousel/carousel"
import type { Video } from "@/lib/api/types"
import YoutubePlayer from "@/components/trailer-modal/YoutubePlayer"
import styles from "./MediaVideosCarousel.module.css"

interface MediaVideosCarouselProps {
    videos: Video[]
    className?: string
}

const VIDEO_TYPE_LABELS: Partial<Record<Video["type"], string>> = {
    Trailer: "Трейлер",
    Clip: "Клип",
}

const getVideoTypeLabel = (type: Video["type"]) => VIDEO_TYPE_LABELS[type] ?? type

const MediaVideosCarousel = ({ videos, className = "" }: MediaVideosCarouselProps) => {
    const thumbnailRefs = useRef<(HTMLButtonElement | null)[]>([])
    const [activeIndex, setActiveIndex] = useState<number | null>(null)

    const youtubeVideos = useMemo(() => {
        const usedKeys = new Set<string>()

        return videos.filter((video) => {
            const isYoutubeVideo = video.site.toLowerCase() === "youtube" && video.key

            if (!isYoutubeVideo || usedKeys.has(video.key)) {
                return false
            }

            usedKeys.add(video.key)
            return true
        })
    }, [videos])

    const activeVideo = activeIndex !== null ? youtubeVideos[activeIndex] : undefined

    useEffect(() => {
        if (!activeVideo) {
            return
        }

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setActiveIndex(null)
                return
            }

            if (event.key === "ArrowLeft") {
                setActiveIndex((index) => index === null ? index : Math.max(0, index - 1))
                return
            }

            if (event.key === "ArrowRight") {
                setActiveIndex((index) => index === null ? index : Math.min(youtubeVideos.length - 1, index + 1))
            }
        }

        document.addEventListener("keydown", handleKeyDown)
        document.body.style.overflow = "hidden"

        return () => {
            document.removeEventListener("keydown", handleKeyDown)
            document.body.style.overflow = ""
        }
    }, [activeVideo, youtubeVideos.length])

    useEffect(() => {
        if (activeIndex === null) {
            return
        }

        thumbnailRefs.current[activeIndex]?.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" })
    }, [activeIndex])

    if (youtubeVideos.length === 0) {
        return null
    }

    return (
        <>
            <Carousel className={`${styles.carousel} ${className}`} title="Видео">
                {youtubeVideos.map((video, index) => (
                    <button
                        key={video.id}
                        type="button"
                        className={styles.videoCard}
                        onClick={() => setActiveIndex(index)}
                        aria-label={`Открыть видео: ${video.name}`}
                    >
                        <span className={styles.preview}>
                            <img
                                src={`https://img.youtube.com/vi/${video.key}/hqdefault.jpg`}
                                alt=""
                                className={styles.previewImage}
                                loading="lazy"
                            />
                            <span className={styles.playIcon} aria-hidden="true">
                                <Play size={24} fill="currentColor" />
                            </span>
                            <span className={styles.badge}>{getVideoTypeLabel(video.type)}</span>
                        </span>
                        <span className={styles.videoName}>{video.name}</span>
                    </button>
                ))}
            </Carousel>

            {activeVideo && (
                <div className={styles.overlay} onClick={() => setActiveIndex(null)}>
                    <div
                        className={styles.modal}
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="media-video-modal-title"
                        onClick={(event) => event.stopPropagation()}
                    >
                        <div className={styles.modalHeader}>
                            <div className={styles.modalTitleGroup}>
                                <p className={styles.modalEyebrow}>{getVideoTypeLabel(activeVideo.type)}</p>
                                <h3 id="media-video-modal-title" className={styles.modalTitle}>{activeVideo.name}</h3>
                            </div>
                            <button type="button" className={styles.closeButton} onClick={() => setActiveIndex(null)} aria-label="Закрыть видео">
                                <X size={20} aria-hidden="true" />
                            </button>
                        </div>

                        <div className={styles.player}>
                            <YoutubePlayer key={activeVideo.key} videoKey={activeVideo.key} title={activeVideo.name} />
                        </div>

                        {youtubeVideos.length > 1 && (
                            <div className={styles.modalCarousel}>
                                <button
                                    type="button"
                                    className={styles.modalArrowButton}
                                    onClick={() => setActiveIndex((index) => index === null ? index : Math.max(0, index - 1))}
                                    disabled={activeIndex === 0}
                                    aria-label="Предыдущее видео"
                                >
                                    <ChevronLeft size={22} aria-hidden="true" />
                                </button>

                                <div className={styles.thumbnails}>
                                    {youtubeVideos.map((video, index) => (
                                        <button
                                            key={video.id}
                                            type="button"
                                            className={`${styles.thumbnail} ${index === activeIndex ? styles.thumbnailActive : ""}`}
                                            ref={(element) => {
                                                thumbnailRefs.current[index] = element
                                            }}
                                            onClick={() => setActiveIndex(index)}
                                            aria-label={`Переключиться на видео: ${video.name}`}
                                            aria-current={index === activeIndex}
                                        >
                                            <img src={`https://img.youtube.com/vi/${video.key}/mqdefault.jpg`} alt="" className={styles.thumbnailImage} />
                                            <span>{index + 1}</span>
                                        </button>
                                    ))}
                                </div>

                                <button
                                    type="button"
                                    className={styles.modalArrowButton}
                                    onClick={() => setActiveIndex((index) => index === null ? index : Math.min(youtubeVideos.length - 1, index + 1))}
                                    disabled={activeIndex === youtubeVideos.length - 1}
                                    aria-label="Следующее видео"
                                >
                                    <ChevronRight size={22} aria-hidden="true" />
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </>
    )
}

export default MediaVideosCarousel