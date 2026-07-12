"use client"

import { useEffect, useState } from "react"
import Carousel from "@/components/carousel/carousel"
import MediaCard from "@/components/media-card/MediaCard"
import { getRecommendations } from "@/lib/api/media"
import type { Media, MediaType } from "@/lib/api/types"

interface MediaRecommendationsCarouselProps {
  mediaId: number
  mediaType: MediaType
  className?: string
}

const getTitle = (mediaType: MediaType) => (mediaType === "tv" ? "Похожие сериалы" : "Похожие кинокартины")

const MediaRecommendationsCarousel = ({ mediaId, mediaType, className }: MediaRecommendationsCarouselProps) => {
  const [recommendations, setRecommendations] = useState<Media[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let isMounted = true

    const loadRecommendations = async () => {
      setIsLoading(true)

      try {
        const response = await getRecommendations(mediaType, mediaId)

        if (!isMounted) return

        setRecommendations(response.results)
      } catch {
        if (!isMounted) return

        setIsLoading(false)
        return
      }

      setIsLoading(false)
    }

    void loadRecommendations()

    return () => {
      isMounted = false
    }
  }, [mediaId, mediaType])

  const title = getTitle(mediaType)

  return (
    <Carousel className={className} title={title}>
      {isLoading ?
        Array.from({ length: 10 }, (_, index) => (
          <MediaCard key={index} isLoading={true} />
        )) :
        recommendations.map((media) => (
          <MediaCard key={media.id} {...media} type={mediaType} />
        ))}
    </Carousel>
  )
}

export default MediaRecommendationsCarousel
