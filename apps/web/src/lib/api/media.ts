import { fetchApi } from "./client"
import type {
    Episode,
    EpisodesResponse,
    Media,
    MediaDetails,
    MediasResponse,
    MediaType,
    MovieList,
    MovieStatus,
    Season,
    TvList,
    TvStatus,
    Video,
    WatchStatus,
} from "./types"

interface ApiMedia {
    id: number
    title: string
    overview?: string
    release_date?: string | null
    poster_path?: string
    vote_average: number
    vote_count?: number
    user_status?: WatchStatus
    type?: MediaType
}

interface ApiMediaDetails extends ApiMedia {
    user_status: WatchStatus
    user_rating: number
    genres: { id: number; name: string }[]
    original_language: string
    origin_country: string[]
    original_title: string
    production_countries: string[]
    popularity: number
    status: MovieStatus | TvStatus
    videos: Video[]
    collection_id?: number
    number_of_seasons?: number
    number_of_episodes?: number
    revenue: number
    budget: number
    runtime?: number
    seasons?: ApiSeason[]
    last_episode_release_date?: string | null
    next_episode_release_date?: string | null
    next_episode_eelease_date?: string | null
}

interface ApiMediasResponse {
    results: ApiMedia[]
    total_pages: number
    total_items: number
}

interface ApiSeason {
    id: number
    release_date?: string | null
    episode_count: number
    title: string
    overview?: string
    poster_path?: string
    season_number: number
    vote_average: number
    vote_count: number
    is_watched?: boolean
}

interface ApiEpisode {
    id: number
    release_date?: string | null
    episode_number: number
    title: string
    overview: string
    runtime: number
    season_number: number
    tv_show_id: number
    poster_path?: string
    vote_average: number
    is_watched?: boolean
}

interface ApiEpisodesResponse {
    results?: ApiEpisode[]
    episodes?: ApiEpisode[]
    total_pages: number
    total_items: number
}

interface ApiWatchStatusResponse {
    watch_status?: string
}

interface ApiWatchListResponse {
    medias: Array<{
        media: ApiMedia
        status: WatchStatus
        type?: MediaType
    }>
    total: number
}

interface WatchListOptions {
    mediaType?: MediaType
    page?: number
    perPage?: number
    status?: WatchStatus
    releaseStatus?: string
    releaseYear?: string
    genreId?: string
    sort?: string
}

export const mapMedia = (media: ApiMedia, type?: MediaType): Media => ({
    id: media.id,
    title: media.title,
    overview: media.overview,
    releaseDate: media.release_date ? new Date(media.release_date) : undefined,
    posterPath: media.poster_path,
    voteAverage: media.vote_average,
    voteCount: media.vote_count ?? 0,
    watchStatus: media.user_status,
    type: (media.type ?? type) as MediaType,
})

export const mapMediasResponse = (data: ApiMediasResponse, type?: MediaType): MediasResponse => ({
    results: data.results.map((m) => mapMedia(m, type)),
    totalPages: data.total_pages,
    totalItems: data.total_items,
})

export const mapSeason = (season: ApiSeason): Season => ({
    id: season.id,
    releaseDate: season.release_date ? new Date(season.release_date) : undefined,
    episodeCount: season.episode_count,
    title: season.title,
    overview: season.overview,
    posterPath: season.poster_path,
    seasonNumber: season.season_number,
    voteAverage: season.vote_average,
    voteCount: season.vote_count,
    isWatched: season.is_watched
})

export const mapEpisode = (episode: ApiEpisode): Episode => ({
    id: episode.id,
    releaseDate: episode.release_date ? new Date(episode.release_date) : undefined,
    episodeNumber: episode.episode_number,
    title: episode.title,
    overview: episode.overview,
    runtime: episode.runtime,
    seasonNumber: episode.season_number,
    tvId: episode.tv_show_id,
    posterPath: episode.poster_path,
    voteAverage: episode.vote_average,
    isWatched: episode.is_watched,
})

export const getMediasList = async (type: MediaType, filter: MovieList | TvList, page = 1): Promise<MediasResponse> => {
    const data = await fetchApi<ApiMediasResponse>(`/v1/${type}/${filter}?page=${page}`)

    return mapMediasResponse(data, type)
}

// export const searchMovies = async (query: string, page = 1): Promise<MoviesResponse> => {
//   const params = new URLSearchParams({ q: query, page: String(page) })
//   const data = await fetchApi<ApiMediasResponse>(`/v1/movie/search?${params.toString()}`)

//   return mapMediasResponse(data)
// }

export const getMediaById = async (type: MediaType, id: number): Promise<MediaDetails> => {
    const data = await fetchApi<ApiMediaDetails>(`/v1/${type}/${id}`, { revalidate: 900 })
    const nextEpisodeReleaseDate = data.next_episode_release_date ?? data.next_episode_eelease_date

    return {
        id: data.id,
        title: data.title,
        overview: data.overview,
        releaseDate: data.release_date ? new Date(data.release_date) : undefined,
        posterPath: data.poster_path,
        userStatus: data.user_status,
        userRating: data.user_rating,
        genres: data.genres.map((genre) => ({ id: genre.id, name: genre.name })),
        originalLanguage: data.original_language,
        originCountry: data.origin_country,
        originalTitle: data.original_title,
        productionCountries: data.production_countries,
        popularity: data.popularity,
        status: data.status,
        videos: data.videos.map((video) => ({
            id: video.id,
            key: video.key,
            name: video.name,
            site: video.site,
            size: video.size,
            type: video.type,
        })),
        voteAverage: data.vote_average,
        voteCount: data.vote_count ?? 0,
        collectionId: data.collection_id,
        numberOfEpisodes: data.number_of_episodes,
        numberOfSeasons: data.number_of_seasons,
        revenue: data.revenue,
        budget: data.budget,
        runtime: data.runtime,
        seasons: data.seasons ? data.seasons.map(mapSeason) : undefined,
        type: type,
        lastEpisodeReleaseDate: data.last_episode_release_date ? new Date(data.last_episode_release_date) : undefined,
        nextEpisodeReleaseDate: nextEpisodeReleaseDate ? new Date(nextEpisodeReleaseDate) : undefined
    }
}

export const getRecommendations = async (type: MediaType, id: number, page = 1): Promise<MediasResponse> => {
    const data = await fetchApi<ApiMediasResponse>(`/v1/${type}/${id}/recommendations?page=${page}`)

    return mapMediasResponse(data, type)
}

export const getWatchStatus = async (id: number): Promise<WatchStatus | null> => {
    try {
        const data = await fetchApi<ApiWatchStatusResponse>(`/v1/watch-list/status?media_id=${id}`)

        return data.watch_status as WatchStatus
    }
    catch {
        return null
    }

}

export const getWatchList = async ({
    mediaType,
    page = 1,
    perPage = 20,
    status,
    releaseStatus,
    releaseYear,
    genreId,
    sort,
}: WatchListOptions): Promise<MediasResponse> => {
    const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage),
    })

    if (mediaType) {
        params.set("media_type", mediaType)
    }

    if (status) {
        params.set("status", status)
    }

    if (releaseStatus) {
        params.set("release_status", releaseStatus)
    }

    if (releaseYear) {
        params.set("year", releaseYear)
    }

    if (genreId) {
        params.set("genre_id", genreId)
    }

    if (sort) {
        params.set("sort", sort)
    }

    const data = await fetchApi<ApiWatchListResponse>(`/v1/watch-list?${params.toString()}`)

    return {
        results: data.medias.flatMap((item): Media[] => {
            const type = item.type ?? item.media.type ?? mediaType

            if (!type) {
                return []
            }

            return [{
                ...mapMedia(item.media, type),
                watchStatus: item.status,
                type,
            }]
        }),
        totalItems: data.total,
        totalPages: Math.max(1, Math.ceil(data.total / perPage)),
    }
}

export const setWatchStatus = (mediaId: number, mediaType: MediaType, status: WatchStatus) =>
    fetchApi("/v1/watch-list/status", {
        method: "POST",
        body: {
            id: mediaId,
            watch_status: status,
            media_type: mediaType,
        },
    })

export const removeWatchStatus = (mediaId: number) =>
    fetchApi(`/v1/watch-list/status?media_id=${mediaId}`, {
        method: "DELETE",
    })

export const getTvSeasons = async (tvId: number): Promise<Season[]> => {
    const data = await fetchApi<ApiMediaDetails>(`/v1/tv/${tvId}`)

    return data.seasons?.map(mapSeason) ?? []
}

export const getSeasonEpisodes = async (tvId: number, seasonNumber: number, page = 1): Promise<EpisodesResponse> => {
    const data = await fetchApi<ApiEpisodesResponse>(`/v1/tv/${tvId}/season/${seasonNumber}?page=${page}`)
    const episodes = data.episodes ?? data.results ?? []

    return {
        episodes: episodes.map(mapEpisode),
        totalPages: data.total_pages,
        totalItems: data.total_items,
    }
}

export const setSeasonWatched = (tvId: number, seasonNumber: number, isWatched: boolean) =>
    fetchApi(`/v1/tv/${tvId}/season/${seasonNumber}/watched`, {
        method: isWatched ? "PUT" : "DELETE",
    })

export const setEpisodeWatched = (
    tvId: number,
    seasonNumber: number,
    episodeNumber: number,
    isWatched: boolean,
) =>
    fetchApi(`/v1/tv/${tvId}/season/${seasonNumber}/episode/${episodeNumber}/watched`, {
        method: isWatched ? "PUT" : "DELETE",
    })

export const searchMulti = async (query: string, page = 1): Promise<MediasResponse> => {
    const params = new URLSearchParams({
        query,
        page: String(page),
    })
    const data = await fetchApi<ApiMediasResponse>(`/v1/search/multi?${params.toString()}`)

    return mapMediasResponse(data)
}





