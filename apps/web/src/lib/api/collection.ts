import { fetchApi } from "./client"
import type { CollectionResponse, Media } from "./types"

interface ApiCollectionPart {
  id: number
  title: string
  overview: string
  poster_path?: string
  type: Media["type"]
  release_date?: string | null
  vote_average: number
  vote_count: number
}

interface ApiCollectionResponse {
  id: number
  name: string
  overview: string
  poster_path: string
  parts: ApiCollectionPart[]
}

export const mapCollection = (data: ApiCollectionResponse): CollectionResponse => (
  {
    id: data.id,
    name: data.name,
    overview: data.overview,
    posterPath: data.poster_path,
    parts: data.parts.map((part) => ({
      id: part.id,
      title: part.title,
      overview: part.overview,
      posterPath: part.poster_path,
      type: part.type,
      releaseDate: part.release_date ? new Date(part.release_date) : undefined,
      voteAverage: part.vote_average,
      voteCount: part.vote_count
    })),
  }
)

export const getCollection = async (id: number): Promise<CollectionResponse> => {
  const data = await fetchApi<ApiCollectionResponse>(`/v1/collections/${id}`, {
    revalidate: 900,
  })

  return mapCollection(data)
}
