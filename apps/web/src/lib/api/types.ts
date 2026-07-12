export type WatchStatus = "watched" | "want_to_watch" | "favorite"

export interface LoginRequest {
    email: string
    password: string
}

export interface RegisterRequest {
    email: string
    username: string
    password: string
}

export interface Genre {
    id: number
    name: string
}

export interface Video {
    id: string
    key: string
    name: string
    site: string
    size: number
    type: "Trailer" | "Clip"
}

export type MediaType = "movie" | "tv"

export const getTypeLabel = (type: MediaType) => type === "tv" ? "Сериал" : "Фильм"

export interface Media {
    id: number
    title: string
    overview?: string
    releaseDate?: Date
    posterPath?: string
    watchStatus?: WatchStatus
    voteAverage: number
    voteCount: number
    type: MediaType
}

export interface MediaDetails extends Media {
    userStatus: WatchStatus
    userRating: number
    genres: Genre[]
    originalLanguage: string
    originCountry: string[]
    originalTitle: string
    productionCountries: string[]
    popularity: number
    status: MovieStatus | TvStatus
    videos: Video[]
    collectionId?: number
    numberOfSeasons?: number
    numberOfEpisodes?: number
    revenue: number
    budget: number
    runtime?: number
    seasons?: Season[]
    lastEpisodeReleaseDate?: Date
    nextEpisodeReleaseDate?: Date
}

export interface MediasResponse {
    results: Media[]
    totalPages: number
    totalItems: number
}

// export interface Movie {
//   id: number
//   title: string
//   overview: string
//   releaseDate?: Date
//   posterPath?: string
//   watchStatus?: WatchStatus
//   voteAverage: number
// }

// export interface MovieDetails extends Movie {
//   userStatus: WatchStatus
//   userRating: number
//   genres: Genre[]
//   originalLanguage: string
//   originCountry: string[]
//   originalTitle: string
//   popularity: number
//   status: MovieStatus
//   videos: Video[]
//   collectionId?: number
// }

export type MovieStatus = "Rumored" | "Planned" | "In Production" | "Post Production" | "Released" | "Canceled"

export type MovieList = "now-playing" | "popular" | "top-rated" | "upcoming"

// export interface MoviesResponse {
//   movies: Movie[]
//   totalPages: number
//   totalItems: number
// }

// export interface Tv {
//   id: number
//   title: string
//   overview: string
//   releaseDate?: Date
//   posterPath?: string
//   watchStatus?: WatchStatus
//   voteAverage: number
// }

export interface Season {
    id: number
    releaseDate?: Date
    episodeCount: number
    title: string
    overview?: string
    posterPath?: string
    seasonNumber: number
    voteAverage: number
    voteCount: number
    isWatched?: boolean
}

export interface Episode {
    id: number
    releaseDate?: Date
    episodeNumber: number
    title: string
    overview: string
    runtime: number
    seasonNumber: number
    tvId: number
    posterPath?: string
    voteAverage: number
    isWatched?: boolean
}

export interface EpisodesResponse {
    episodes: Episode[]
    totalPages: number
    totalItems: number
}

// export interface TvDetails extends Tv {
//   userStatus: WatchStatus
//   userRating: number
//   genres: Genre[]
//   originalLanguage: string
//   originCountry: string[]
//   originalTitle: string
//   popularity: number
//   status: TvStatus
//   videos: Video[]
//   numberOfSeasons: number
//   numberOfEpisodes: number
//   seasons: Season[]
// }

export type TvStatus = "Returning Series" | "Planned" | "In Production" | "Ended" | "Canceled" | "Pilot"

export type TvList = "airing-today" | "popular" | "top-rated" | "on-the-air"

// export interface TvResponse {
//   tv: Tv[]
//   totalPages: number
//   totalItems: number
// }

export interface CollectionResponse {
    id: number
    name: string
    overview: string
    posterPath: string
    parts: Media[]
}
