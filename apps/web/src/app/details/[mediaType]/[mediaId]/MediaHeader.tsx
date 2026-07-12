import {Genre, MediaType, MovieStatus, TvStatus} from '@/lib/api/types'
import styles from './MediaHeader.module.css'
import {getSession} from '@/lib/auth/session'
import Rating from '@/components/rating/rating'
import MediaStatusButtons from './MediaStatusButtons'
import {convertNumberToCurrency} from "@/lib/utils/utils";
import {getWatchStatus} from "@/lib/api/media";

export interface MediaHeaderProps {
    id: number
    type: MediaType
    title: string
    originalTitle: string
    overview?: string
    posterPath?: string
    releaseDate?: Date
    lastEpisodeReleaseDate?: Date
    nextEpisodeReleaseDate?: Date
    originalLanguage: string
    originCountry: string[]
    productionCountries: string[]
    revenue: number
    budget?: number
    runtime?: number
    genres: Genre[]
    voteAverage?: number
    voteCount?: number
    numberOfEpisodes?: number
    numberOfSeasons?: number
    status?: MovieStatus | TvStatus
}

const STATUS_LABELS: Record<MovieStatus | TvStatus, string> = {
    "Rumored": "По слухам",
    "Planned": "Запланировано",
    "In Production": "В производстве",
    "Post Production": "Постпродакшн",
    "Released": "Вышел",
    "Canceled": "Отменено",
    "Returning Series": "Продолжается",
    "Ended": "Завершен",
    "Pilot": "Пилот",
}

const getStatusLabel = (status?: MovieStatus | TvStatus) => status ? STATUS_LABELS[status] ?? status : "Неизвестно"

const formatRuntime = (runtime: number) => {
    const hours = Math.floor(runtime / 60)
    const minutes = runtime % 60

    if (hours === 0) {
        return `${minutes} мин.`
    }

    return minutes > 0 ? `${hours} ч ${minutes} мин.` : `${hours} ч`
}

const MediaHeader = async (props: MediaHeaderProps) => {
    const session = await getSession()
    const watchStatus = session.isAuthenticated ? await getWatchStatus(props.id) : null

    return (
        <div className={styles.mainContainer}>
            <div className={styles.posterContainer}>
                <div className={styles.poster}>
                    {props.posterPath ? (
                        <img src={props.posterPath} alt={props.title} className={styles.posterImage}/>
                    ) : (
                        <div className={styles.posterPlaceholder}>🎬</div>
                    )}
                </div>
                {session.isAuthenticated && (
                    <MediaStatusButtons mediaId={props.id} mediaType={props.type} status={watchStatus}/>
                )}
            </div>
            <div className={styles.info}>
                <div className={styles.titleContainer}>
                    <p className={styles.mediaType}>{props.type === "tv" ? "Сериал" : "Фильм"}</p>
                    <h1 className={styles.title}>{props.title}</h1>
                    <div className={styles.quickMeta}>
                        {props.genres.slice(0, 3).map((genre) => (
                            <span key={genre.id}>{genre.name}</span>
                        ))}
                    </div>
                </div>
                <table className={styles.table}>
                    <tbody>
                    <tr>
                        <td>Оригинальное название</td>
                        <td>{props.originalTitle}</td>
                    </tr>
                    <tr>
                        <td>Оценка</td>
                        <td>
                            <Rating voteAverage={props.voteAverage} voteCount={props.voteCount}/>
                        </td>
                    </tr>
                    {props.releaseDate && (
                        <tr>
                            <td>Дата выхода</td>
                            <td>{props.releaseDate.toLocaleDateString('ru')}</td>
                        </tr>
                    )}
                    {props.lastEpisodeReleaseDate && (
                        <tr>
                            <td>Последний эпизод</td>
                            <td>{props.lastEpisodeReleaseDate.toLocaleDateString('ru')}</td>
                        </tr>
                    )}
                    {props.nextEpisodeReleaseDate && (
                        <tr>
                            <td>Следующий эпизод</td>
                            <td>{props.nextEpisodeReleaseDate.toLocaleDateString('ru')}</td>
                        </tr>
                    )}
                    {props.genres.length > 0 && (
                        <tr>
                            <td>Жанр</td>
                            <td>{props.genres.map((g) => g.name).join(", ")}</td>
                        </tr>
                    )}
                    {props.productionCountries.length > 0 && (
                        <tr>
                            <td>Страна производства</td>
                            <td>{props.productionCountries.join(", ")}</td>
                        </tr>
                    )}
                    {props.budget != undefined && props.budget !== 0 && (
                        <tr>
                            <td>Бюджет</td>
                            <td>{convertNumberToCurrency(props.budget, 'USD', "ru")}</td>
                        </tr>
                    )}
                    {props.revenue != undefined && props.revenue !== 0 && (
                        <tr>
                            <td>Сборы</td>
                            <td>{convertNumberToCurrency(props.revenue, 'USD', 'ru')}</td>
                        </tr>
                    )}
                    {props.runtime != undefined && props.runtime !== 0 && (
                        <tr>
                            <td>Длительность</td>
                            <td>{formatRuntime(props.runtime)}</td>
                        </tr>
                    )}
                    <tr>
                        <td>Статус</td>
                        <td>{getStatusLabel(props.status)}</td>
                    </tr>
                    {props.numberOfSeasons && (
                        <tr>
                            <td>Количество сезонов</td>
                            <td>{props.numberOfSeasons}</td>
                        </tr>
                    )}
                    {props.numberOfEpisodes && (
                        <tr>
                            <td>Количество серий</td>
                            <td>{props.numberOfEpisodes}</td>
                        </tr>
                    )}
                    </tbody>
                </table>

                {props.overview && <p className={styles.overview}>{props.overview}</p>}
            </div>
        </div>
    )
}

export default MediaHeader