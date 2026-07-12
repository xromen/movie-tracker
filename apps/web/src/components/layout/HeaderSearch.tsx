"use client"

import Link from "next/link"
import { useRouter } from "next/navigation"
import { useEffect, useRef, useState } from "react"
import type { KeyboardEvent as ReactKeyboardEvent } from "react"
import { Search, X } from "lucide-react"
import { searchMulti } from "@/lib/api/media"
import {getTypeLabel, Media} from "@/lib/api/types"
import styles from "./HeaderSearch.module.css"

const SEARCH_DELAY_MS = 350
const MIN_QUERY_LENGTH = 2
const VISIBLE_RESULTS = 5
const RESULTS_LIST_ID = "header-search-results"

const formatDate = (date?: Date) => date?.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
})

const getRatingClassName = (voteAverage: number) => {
    if (voteAverage >= 7) return styles.ratingGood
    if (voteAverage >= 5) return styles.ratingMedium
    return styles.ratingLow
}

const getResultId = (media: Media) => `header-search-result-${media.type}-${media.id}`

const HeaderSearch = () => {
    const router = useRouter()
    const [query, setQuery] = useState("")
    const [results, setResults] = useState<Media[]>([])
    const [totalItems, setTotalItems] = useState(0)
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState(false)
    const [isOpen, setIsOpen] = useState(false)
    const [activeIndex, setActiveIndex] = useState(-1)
    const rootRef = useRef<HTMLDivElement>(null)
    const normalizedQuery = query.trim()
    const shouldSearch = normalizedQuery.length >= MIN_QUERY_LENGTH
    const activeResult = activeIndex >= 0 ? results[activeIndex] : undefined

    useEffect(() => {
        const handlePointerDown = (event: PointerEvent) => {
            if (!rootRef.current?.contains(event.target as Node)) {
                setIsOpen(false)
                setActiveIndex(-1)
            }
        }

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setIsOpen(false)
                setActiveIndex(-1)
            }
        }

        document.addEventListener("pointerdown", handlePointerDown)
        document.addEventListener("keydown", handleKeyDown)

        return () => {
            document.removeEventListener("pointerdown", handlePointerDown)
            document.removeEventListener("keydown", handleKeyDown)
        }
    }, [])

    useEffect(() => {
        if (!shouldSearch) {
            return
        }

        let isCurrent = true

        const timeoutId = window.setTimeout(async () => {
            try {
                const response = await searchMulti(normalizedQuery, 1)

                if (!isCurrent) return

                setResults(response.results.slice(0, VISIBLE_RESULTS))
                setTotalItems(response.totalItems)
            } catch {
                if (!isCurrent) return

                setResults([])
                setTotalItems(0)
                setError(true)
            } finally {
                if (isCurrent) {
                    setIsLoading(false)
                }
            }
        }, SEARCH_DELAY_MS)

        return () => {
            isCurrent = false
            window.clearTimeout(timeoutId)
        }
    }, [normalizedQuery, shouldSearch])

    const handleQueryChange = (value: string) => {
        setQuery(value)
        setActiveIndex(-1)

        if (value.trim().length >= MIN_QUERY_LENGTH) {
            setIsOpen(true)
            setIsLoading(true)
            setError(false)
        } else {
            setResults([])
            setTotalItems(0)
            setIsLoading(false)
            setError(false)
        }
    }

    const clearSearch = () => {
        setQuery("")
        setResults([])
        setTotalItems(0)
        setIsOpen(false)
        setError(false)
        setActiveIndex(-1)
    }

    const navigateToResult = (media: Media) => {
        setIsOpen(false)
        setActiveIndex(-1)
        router.push(`/details/${media.type}/${media.id}`)
    }

    const handleInputKeyDown = (event: ReactKeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Escape") {
            clearSearch()
            event.preventDefault()
            return
        }

        if (!shouldSearch) {
            return
        }

        if (event.key === "ArrowDown") {
            event.preventDefault()
            setIsOpen(true)
            setActiveIndex((index) => (results.length === 0 ? -1 : (index + 1) % results.length))
            return
        }

        if (event.key === "ArrowUp") {
            event.preventDefault()
            setIsOpen(true)
            setActiveIndex((index) => {
                if (results.length === 0) return -1
                return index <= 0 ? results.length - 1 : index - 1
            })
            return
        }

        if (event.key === "Enter" && activeResult) {
            event.preventDefault()
            navigateToResult(activeResult)
        }
    }

    return (
        <div className={styles.search} ref={rootRef}>
            <label className="visuallyHidden" htmlFor="header-search">
                Поиск фильмов и сериалов
            </label>
            <div className={`${styles.inputWrap} ${isOpen && shouldSearch ? styles.inputWrapActive : ""}`}>
                <Search className={styles.searchIcon} size={18} aria-hidden="true" />
                <input
                    id="header-search"
                    className={styles.input}
                    value={query}
                    onChange={(event) => handleQueryChange(event.target.value)}
                    onFocus={() => shouldSearch && setIsOpen(true)}
                    onKeyDown={handleInputKeyDown}
                    placeholder="Поиск"
                    type="search"
                    autoComplete="off"
                    role="combobox"
                    aria-autocomplete="list"
                    aria-expanded={isOpen && shouldSearch}
                    aria-controls={RESULTS_LIST_ID}
                    aria-activedescendant={activeResult ? getResultId(activeResult) : undefined}
                />
                {query && (
                    <button className={styles.clearButton} type="button" onClick={clearSearch} aria-label="Очистить поиск">
                        <X size={18} aria-hidden="true" />
                    </button>
                )}
            </div>

            {isOpen && shouldSearch && (
                <div className={styles.modal}>
                    {isLoading ? (
                        <div className={styles.resultsList} aria-label="Загрузка результатов поиска">
                            {Array.from({ length: VISIBLE_RESULTS }, (_, index) => (
                                <SearchResultSkeleton key={index} />
                            ))}
                        </div>
                    ) : error ? (
                        <p className={styles.stateText} role="status">Не удалось выполнить поиск</p>
                    ) : results.length > 0 ? (
                        <div className={styles.resultsList} id={RESULTS_LIST_ID} role="listbox" aria-label="Результаты поиска">
                            {results.map((media, index) => (
                                <SearchResultItem
                                    key={`${media.type}-${media.id}`}
                                    id={getResultId(media)}
                                    media={media}
                                    isActive={index === activeIndex}
                                    onActive={() => setActiveIndex(index)}
                                    onNavigate={() => setIsOpen(false)}
                                />
                            ))}
                        </div>
                    ) : (
                        <p className={styles.stateText} role="status">Ничего не найдено</p>
                    )}

                    {!isLoading && totalItems > VISIBLE_RESULTS && (
                        <span className={styles.totalText}>Найдено: {totalItems}</span>
                    )}
                </div>
            )}
        </div>
    )
}

interface SearchResultItemProps {
    id: string
    media: Media
    isActive: boolean
    onActive: () => void
    onNavigate: () => void
}

const SearchResultItem = ({ id, media, isActive, onActive, onNavigate }: SearchResultItemProps) => {
    const shouldShowRating = media.voteAverage > 0

    return (
        <Link
            id={id}
            className={`${styles.result} ${isActive ? styles.resultActive : ""}`}
            href={`/details/${media.type}/${media.id}`}
            onClick={onNavigate}
            onMouseEnter={onActive}
            role="option"
            aria-selected={isActive}
        >
            <div className={styles.poster}>
                {media.posterPath ? (
                    <img src={media.posterPath} alt={media.title} className={styles.posterImage} />
                ) : (
                    <span className={styles.posterPlaceholder}>Нет постера</span>
                )}
            </div>
            <div className={styles.resultInfo}>
                <div className={styles.resultTopline}>
                    <h3 className={styles.resultTitle}>{media.title}</h3>
                    <span className={styles.typeBadge}>{getTypeLabel(media.type)}</span>
                </div>
                <div className={styles.metaLine}>
                    {shouldShowRating && (
                        <span className={`${styles.rating} ${getRatingClassName(media.voteAverage)}`}>
                            {media.voteAverage.toFixed(1)}
                        </span>
                    )}
                    {media.releaseDate && <span>{formatDate(media.releaseDate)}</span>}
                </div>
                {media.overview && <p className={styles.overview}>{media.overview}</p>}
            </div>
        </Link>
    )
}

const SearchResultSkeleton = () => (
    <div className={styles.skeletonResult} aria-hidden="true">
        <div className={`${styles.skeleton} ${styles.skeletonPoster}`} />
        <div className={styles.skeletonInfo}>
            <div className={`${styles.skeleton} ${styles.skeletonTitle}`} />
            <div className={`${styles.skeleton} ${styles.skeletonMeta}`} />
            <div className={`${styles.skeleton} ${styles.skeletonText}`} />
        </div>
    </div>
)

export default HeaderSearch