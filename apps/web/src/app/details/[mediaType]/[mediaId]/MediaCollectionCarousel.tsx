"use client"

import { useEffect, useState } from "react"
import Carousel from "@/components/carousel/carousel"
import MediaCard from "@/components/media-card/MediaCard"
import { getCollection } from "@/lib/api/collection"
import type { CollectionResponse } from "@/lib/api/types"

interface MediaCollectionCarouselProps {
  collectionId: number
  className?: string
}

const MediaCollectionCarousel = ({ collectionId, className }: MediaCollectionCarouselProps) => {
  const [collection, setCollection] = useState<CollectionResponse>()
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let isMounted = true

    const loadCollection = async () => {
      setIsLoading(true)

      try {
        const loadedCollection = await getCollection(collectionId)

        if (!isMounted) return

        setCollection(loadedCollection)
      } catch {
        if (!isMounted) return

        setIsLoading(false)
        return
      }

      setIsLoading(false)
    }

    void loadCollection()

    return () => {
      isMounted = false
    }
  }, [collectionId])

  if (isLoading) {
    return (
      <Carousel className={className} title="Загрузка...">
        {
          Array.from({ length: 10 }, (_, index) => (
            <MediaCard key={index} isLoading={true} />
          ))
        }
      </Carousel>
    )
  }

  if (!collection || collection.parts.length === 0) {
    return null
  }

  const parts = collection.parts
    .slice()
    .sort((a, b) => {
      if (!a.releaseDate && !b.releaseDate) return 0
      if (!a.releaseDate) return 1
      if (!b.releaseDate) return -1
      return a.releaseDate.getTime() - b.releaseDate.getTime()
    })

  return (
    <Carousel className={className} title={collection.name}>
      {parts.map((movie) => (
        <MediaCard key={movie.id} {...movie} type="movie" />
      ))}
    </Carousel>
  )
}

export default MediaCollectionCarousel
