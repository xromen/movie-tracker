"use client"

import Link from "next/link"
import {Bookmark, Check, Heart, Info, Loader2} from "lucide-react"
import {
    type FocusEvent as ReactFocusEvent,
    useCallback,
    useEffect,
    useId,
    useLayoutEffect,
    useRef,
    useState,
    useTransition,
} from "react"
import {createPortal} from "react-dom"
import ContentLoader from "react-content-loader"
import {removeWatchStatus, setWatchStatus} from "@/lib/api/media"
import {getTypeLabel, MediaType, WatchStatus} from "@/lib/api/types"
import styles from "./MediaCard.module.css"

interface MediaCardProps {
    id?: number
    title?: string
    overview?: string
    releaseDate?: Date
    posterPath?: string
    className?: string
    voteAverage?: number
    voteCount?: number
    watchStatus?: WatchStatus
    type?: MediaType
    isLoading?: boolean
    withTypeBadge?: boolean
}

const POPOVER_WIDTH = 340
const POPOVER_GAP = 12
const VIEWPORT_MARGIN = 16

const STATUS_OPTIONS: {
    status: WatchStatus
    label: string
    Icon: typeof Check
}[] = [
    {status: "watched", label: "Просмотрено", Icon: Check},
    {status: "want_to_watch", label: "Хочу посмотреть", Icon: Bookmark},
    {status: "favorite", label: "Избранное", Icon: Heart},
]

const getColor = (value?: number): string => {
    if (!value) return ""
    if (value >= 7) return "#1D9E75"
    if (value >= 5) return "#EF9F27"
    return "#E24B4A"
}

const formatDate = (date?: Date) => date?.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
})

const formatVotes = (count?: number) => {
    if (!count) return null

    return count >= 1000 ? `${(count / 1000).toFixed(1)}К оценки` : `${count} оценки`
}

const MediaCard = ({
                       id,
                       title,
                       overview,
                       releaseDate,
                       posterPath,
                       className = "",
                       voteAverage,
                       voteCount,
                       watchStatus,
                       type,
                       isLoading,
                       withTypeBadge,
                   }: MediaCardProps) => {
    const popoverId = useId()
    const infoButtonRef = useRef<HTMLButtonElement>(null)
    const popoverRef = useRef<HTMLDivElement>(null)
    const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
    const [activeStatus, setActiveStatus] = useState<WatchStatus | undefined>(watchStatus)
    const [pendingStatus, setPendingStatus] = useState<WatchStatus | null>(null)
    const [error, setError] = useState("")
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const [popoverPosition, setPopoverPosition] = useState({top: 0, left: 0})
    const [isPending, startTransition] = useTransition()

    const updatePopoverPosition = useCallback(() => {
        const button = infoButtonRef.current

        if (!button) return

        const rect = button.getBoundingClientRect()
        const popoverWidth = Math.min(POPOVER_WIDTH, window.innerWidth - VIEWPORT_MARGIN * 2)
        const measuredHeight = popoverRef.current?.offsetHeight ?? 260
        const spaceRight = window.innerWidth - rect.right - VIEWPORT_MARGIN
        const canOpenRight = spaceRight >= popoverWidth + POPOVER_GAP
        const left = canOpenRight
            ? rect.right + POPOVER_GAP
            : Math.max(VIEWPORT_MARGIN, rect.left - popoverWidth - POPOVER_GAP)
        const maxTop = Math.max(VIEWPORT_MARGIN, window.innerHeight - measuredHeight - VIEWPORT_MARGIN)
        const top = Math.min(Math.max(VIEWPORT_MARGIN, rect.top), maxTop)

        setPopoverPosition({top, left})
    }, [])

    const clearCloseTimer = useCallback(() => {
        if (closeTimerRef.current) {
            clearTimeout(closeTimerRef.current)
            closeTimerRef.current = null
        }
    }, [])

    const openPopover = useCallback(() => {
        clearCloseTimer()
        updatePopoverPosition()
        setIsPopoverOpen(true)
    }, [clearCloseTimer, updatePopoverPosition])

    const scheduleClosePopover = useCallback(() => {
        clearCloseTimer()
        closeTimerRef.current = setTimeout(() => setIsPopoverOpen(false), 120)
    }, [clearCloseTimer])

    const handleBlur = (event: ReactFocusEvent<HTMLElement>) => {
        const nextTarget = event.relatedTarget as Node | null

        if (
            nextTarget &&
            (infoButtonRef.current?.contains(nextTarget) || popoverRef.current?.contains(nextTarget))
        ) {
            return
        }

        setIsPopoverOpen(false)
    }

    useEffect(() => () => clearCloseTimer(), [clearCloseTimer])

    useEffect(() => {
        if (!isPopoverOpen) return

        window.addEventListener("resize", updatePopoverPosition)
        window.addEventListener("scroll", updatePopoverPosition, true)

        return () => {
            window.removeEventListener("resize", updatePopoverPosition)
            window.removeEventListener("scroll", updatePopoverPosition, true)
        }
    }, [isPopoverOpen, updatePopoverPosition])

    useLayoutEffect(() => {
        if (isPopoverOpen) {
            updatePopoverPosition()
        }
    }, [isPopoverOpen, updatePopoverPosition])

    if (isLoading) {
        return SkeletonCard(className)
    }

    if (!id || !title || !type) {
        return null
    }

    const shouldShowRating = voteAverage !== undefined && voteAverage !== null && voteAverage !== 0
    const ratingColor = getColor(voteAverage)
    const releaseLabel = formatDate(releaseDate)
    const votesLabel = formatVotes(voteCount)

    const handleStatusClick = (status: WatchStatus) => {
        const nextStatus = activeStatus === status ? undefined : status
        const previousStatus = activeStatus

        setActiveStatus(nextStatus)
        setPendingStatus(status)
        setError("")

        startTransition(async () => {
            try {
                if (nextStatus) {
                    await setWatchStatus(id, type, nextStatus)
                } else {
                    await removeWatchStatus(id)
                }
            } catch {
                setActiveStatus(previousStatus)
                setError("Не удалось обновить отметку")
            } finally {
                setPendingStatus(null)
            }
        })
    }

    const popover = isPopoverOpen && typeof document !== "undefined" ? createPortal(
        <div
            ref={popoverRef}
            id={popoverId}
            className={styles.popover}
            role="dialog"
            aria-label={`Быстрые действия: ${title}`}
            style={{top: popoverPosition.top, left: popoverPosition.left}}
            onPointerEnter={openPopover}
            onPointerLeave={scheduleClosePopover}
            onFocus={openPopover}
            onBlur={handleBlur}
        >
            <div className={styles.popoverInfo}>
                <div className={styles.popoverHeader}>
                    <div className={styles.popoverTitleGroup}>
                        <span className={styles.mediaType}>{getTypeLabel(type)}</span>
                        <h4 className={styles.popoverTitle}>{title}</h4>
                    </div>

                    {shouldShowRating && (
                        <span className={styles.popoverRating} style={{background: ratingColor}}>
              {voteAverage.toFixed(1)}
            </span>
                    )}
                </div>

                <div className={styles.popoverMeta}>
                    {releaseLabel && <span>{releaseLabel}</span>}
                    {votesLabel && <span>{votesLabel}</span>}
                </div>

                {overview && <p className={styles.overview}>{overview}</p>}

                <div className={styles.statusButtons} aria-label="Статус просмотра">
                    {STATUS_OPTIONS.map(({status, label, Icon}) => {
                        const isActive = activeStatus === status
                        const isLoadingStatus = isPending && pendingStatus === status

                        return (
                            <button
                                key={status}
                                className={`${styles.statusButton} ${isActive ? styles.statusButtonActive : ""}`}
                                type="button"
                                onClick={() => handleStatusClick(status)}
                                disabled={isPending}
                                aria-pressed={isActive}
                                title={isActive ? `Снять отметку: ${label}` : label}
                            >
                                {isLoadingStatus ? (
                                    <Loader2 className={styles.spinner} size={15} aria-hidden="true"/>
                                ) : (
                                    <Icon size={15} aria-hidden="true"/>
                                )}
                                <span>{label}</span>
                            </button>
                        )
                    })}
                </div>

                {error && <p className={styles.error} role="alert">{error}</p>}
            </div>
        </div>,
        document.body,
    ) : null

    return (
        <article className={`${className} ${styles.card}`}>
            <Link href={`/details/${type}/${id}`} className={styles.cardLink}>
                <div className={styles.poster}>
                    {posterPath ? (
                        <img src={posterPath} alt={title} className={styles.posterImage}/>
                    ) : (
                        <div className={styles.posterPlaceholder}>Нет постера</div>
                    )}
                </div>

                <div className={styles.info}>
                    <h3 className={styles.title}>{title}</h3>
                    <div className={styles.meta}>
                        <span className={styles.releaseDate}>{releaseLabel}</span>

                        {shouldShowRating && (
                            <div className={styles.rating} style={{background: ratingColor}}>
                                {voteAverage.toFixed(1)}
                            </div>
                        )}
                    </div>
                </div>
            </Link>

            {withTypeBadge && (
                <span className={styles.typeBadge}>{getTypeLabel(type)}</span>
            )}

            <div
                className={styles.quickActions}
                onPointerEnter={openPopover}
                onPointerLeave={scheduleClosePopover}
                onFocus={openPopover}
                onBlur={handleBlur}
            >
                <button
                    ref={infoButtonRef}
                    className={styles.infoButton}
                    type="button"
                    aria-label={`Информация о медиа: ${title}`}
                    aria-haspopup="dialog"
                    aria-expanded={isPopoverOpen}
                    aria-controls={isPopoverOpen ? popoverId : undefined}
                >
                    <Info size={16} aria-hidden="true"/>
                </button>
            </div>

            {popover}
        </article>
    )
}

const SkeletonCard = (className: string) => (
    <div className={`${className} ${styles.card}`}>
        <div className={styles.poster}>
            <ContentLoader
                width="100%"
                height="100%"
                viewBox="0 0 100 100"
                preserveAspectRatio="none"
                backgroundColor="#6d6d6d8c"
                foregroundColor="#8d8d8d83"
            >
                <rect x="0" y="0" width="100" height="100"/>
            </ContentLoader>
        </div>

        <div className={styles.infoSkeleton}>
            <ContentLoader
                width="100%"
                className={styles.titleSkeleton}
                viewBox="0 0 100 100"
                preserveAspectRatio="none"
                backgroundColor="#6d6d6d8c"
                foregroundColor="#8d8d8d83"
            >
                <rect x="0" y="0" width="100" height="100"/>
            </ContentLoader>
            <ContentLoader
                width="70%"
                className={styles.metaSkeleton}
                viewBox="0 0 100 100"
                preserveAspectRatio="none"
                backgroundColor="#6d6d6d8c"
                foregroundColor="#8d8d8d83"
            >
                <rect x="0" y="0" width="100" height="100"/>
            </ContentLoader>
            <ContentLoader
                width="90%"
                className={styles.metaSkeleton}
                viewBox="0 0 100 100"
                preserveAspectRatio="none"
                backgroundColor="#6d6d6d8c"
                foregroundColor="#8d8d8d83"
            >
                <rect x="0" y="0" width="100" height="100"/>
            </ContentLoader>
        </div>
    </div>
)

export default MediaCard
