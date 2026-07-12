"use client";

import {CalendarDays, Check, ChevronDown, Clock3, Film, Heart, ListFilter, SortAsc, Star, Tv, X} from "lucide-react";
import {useRouter, useSearchParams} from "next/navigation";
import {useEffect, useMemo, useState, useTransition} from "react";
import MediaCard from "@/components/media-card/MediaCard";
import Pagination from "@/components/pagination/Pagination";
import {getWatchList} from "@/lib/api/media";
import type {Media, MediaType, WatchStatus} from "@/lib/api/types";
import styles from "./WatchlistPage.module.css";

type StatusTabKey = "all" | WatchStatus;
type QueryFilterKey = "media_type" | "release_status" | "year" | "genre_id";

type FilterOption = {
    label: string;
    value?: string;
};

type FilterConfig = {
    key: QueryFilterKey;
    label: string;
    options: FilterOption[];
};

const STATUS_TABS: { label: string; value?: WatchStatus; key: StatusTabKey; Icon: typeof ListFilter }[] = [
    {label: "Все", key: "all", Icon: ListFilter},
    {label: "Просмотрено", value: "watched", key: "watched", Icon: Check},
    {label: "Хочу посмотреть", value: "want_to_watch", key: "want_to_watch", Icon: Clock3},
    {label: "Избранное", value: "favorite", key: "favorite", Icon: Heart},
];

const FILTERS: FilterConfig[] = [
    {
        key: "media_type",
        label: "Тип",
        options: [
            {label: "Все типы"},
            {label: "Фильм", value: "movie"},
            {label: "Сериал", value: "tv"},
        ],
    },
    {
        key: "release_status",
        label: "Статус релиза",
        options: [
            {label: "Любой статус"},
            {label: "Вышло", value: "Released"},
            {label: "Продолжается", value: "Returning Series"},
            {label: "В производстве", value: "In Production"},
            {label: "Запланировано", value: "Planned"},
            {label: "Завершено", value: "Ended"},
            {label: "Отменено", value: "Canceled"},
        ],
    },
    {
        key: "year",
        label: "Год выпуска",
        options: [
            {label: "Любой год"},
            ...Array.from({length: 16}, (_, index) => {
                const year = String(new Date().getFullYear() - index);

                return {label: year, value: year};
            }),
        ],
    },
    {
        key: "genre_id",
        label: "Жанр",
        options: [
            {label: "Любой жанр"},
            {label: "Боевик", value: "28"},
            {label: "Приключения", value: "12"},
            {label: "Анимация", value: "16"},
            {label: "Комедия", value: "35"},
            {label: "Драма", value: "18"},
            {label: "Фэнтези", value: "14"},
            {label: "Ужасы", value: "27"},
            {label: "Детектив", value: "9648"},
            {label: "Фантастика", value: "878"},
            {label: "Триллер", value: "53"},
        ],
    },
];

const SORT_OPTIONS: FilterOption[] = [
    {label: "По дате добавления", value: "added_at_desc"},
    {label: "По дате добавления", value: "added_at_asc"},
    {label: "По дате обновления", value: "updated_at_desc"},
    {label: "По дате обновления", value: "updated_at_asc"},
    {label: "По оценке пользователя", value: "user_rating_desc"},
    {label: "По оценке пользователя", value: "user_rating_asc"},
];

const isMediaType = (value: string | null): value is MediaType => value === "movie" || value === "tv";
const isWatchStatus = (value: string | null): value is WatchStatus => (
    value === "watched" || value === "want_to_watch" || value === "favorite"
);

const getOptionLabel = (options: FilterOption[], value?: string) => (
    options.find((option) => option.value === value)?.label ?? options[0]?.label ?? ""
);

const getSortIcon = (value?: string) => value?.endsWith("_asc") ? "↧" : "↥";

const WatchlistClient = () => {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [items, setItems] = useState<Media[]>([]);
    const [totalPages, setTotalPages] = useState(1);
    const [totalItems, setTotalItems] = useState(0);
    const [statusCounts, setStatusCounts] = useState<Record<StatusTabKey, number>>({
        all: 0,
        watched: 0,
        want_to_watch: 0,
        favorite: 0,
    });
    const [openMenu, setOpenMenu] = useState<string | null>(null);
    const [isLoading, startTransition] = useTransition();
    const [isError, setIsError] = useState(false);

    const activeMediaType = isMediaType(searchParams.get("media_type")) ? searchParams.get("media_type") as MediaType : undefined;
    const activeStatus = isWatchStatus(searchParams.get("status")) ? searchParams.get("status") as WatchStatus : undefined;
    const releaseStatus = searchParams.get("release_status") ?? undefined;
    const releaseYear = searchParams.get("year") ?? undefined;
    const genreId = searchParams.get("genre_id") ?? undefined;
    const activeSort = searchParams.get("sort") ?? undefined;
    const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
    const perPage = 20;

    const activeFilterCount = [activeMediaType, releaseStatus, releaseYear, genreId].filter(Boolean).length;

    useEffect(() => {
        let isActive = true;

        startTransition(async () => {
            setIsError(false);

            try {
                const response = await getWatchList({
                    mediaType: activeMediaType,
                    page,
                    perPage,
                    status: activeStatus,
                    releaseStatus,
                    releaseYear,
                    genreId,
                    sort: activeSort,
                });

                if (!isActive) return;

                setItems(response.results);
                setTotalPages(response.totalPages);
                setTotalItems(response.totalItems);
            } catch {
                if (isActive) {
                    setIsError(true);
                }
            }
        });

        return () => {
            isActive = false;
        };
    }, [activeMediaType, activeSort, activeStatus, genreId, page, releaseStatus, releaseYear]);

    useEffect(() => {
        let isActive = true;

        const loadCounts = async () => {
            try {
                const responses = await Promise.all(
                    STATUS_TABS.map((tab) => getWatchList({
                        mediaType: activeMediaType,
                        page: 1,
                        perPage: 1,
                        status: tab.value,
                        releaseStatus,
                        releaseYear,
                        genreId,
                    })),
                );

                if (!isActive) return;

                setStatusCounts(
                    STATUS_TABS.reduce<Record<StatusTabKey, number>>((acc, tab, index) => {
                        acc[tab.key] = responses[index]?.totalItems ?? 0;

                        return acc;
                    }, {all: 0, watched: 0, want_to_watch: 0, favorite: 0}),
                );
            } catch {
                if (isActive) {
                    setStatusCounts((current) => ({...current, [activeStatus ?? "all"]: totalItems}));
                }
            }
        };

        void loadCounts();

        return () => {
            isActive = false;
        };
    }, [activeMediaType, activeStatus, genreId, releaseStatus, releaseYear, totalItems]);

    const searchParamValues = useMemo(() => ({
        media_type: activeMediaType,
        status: activeStatus,
        release_status: releaseStatus,
        year: releaseYear,
        genre_id: genreId,
        sort: activeSort,
    }), [activeMediaType, activeSort, activeStatus, genreId, releaseStatus, releaseYear]);

    const replaceParams = (updater: (params: URLSearchParams) => void) => {
        const params = new URLSearchParams(searchParams.toString());

        updater(params);

        const query = params.toString();
        router.replace(query ? `/watchlist?${query}` : "/watchlist");
    };

    const updateParam = (key: string, value?: string) => {
        replaceParams((params) => {
            if (value) {
                params.set(key, value);
            } else {
                params.delete(key);
            }

            params.delete("page");
        });
        setOpenMenu(null);
    };

    const clearFilters = () => {
        replaceParams((params) => {
            FILTERS.forEach((filter) => params.delete(filter.key));
            params.delete("page");
        });
        setOpenMenu(null);
    };

    const selectedSort = activeSort ?? SORT_OPTIONS[0].value;
    const activeSortLabel = getOptionLabel(SORT_OPTIONS, selectedSort);

    return (
        <section className={styles.page}>
            <div className={styles.statusTabs} role="tablist" aria-label="Статус просмотра">
                {STATUS_TABS.map(({label, value, key, Icon}) => {
                    const isActive = activeStatus === value;

                    return (
                        <button
                            key={key}
                            className={`${styles.statusTab} ${isActive ? styles.statusTabActive : ""}`}
                            type="button"
                            role="tab"
                            aria-selected={isActive}
                            onClick={() => updateParam("status", value)}
                        >
                            <Icon size={15} aria-hidden="true"/>
                            <span>{label}</span>
                            <span className={styles.count}>{statusCounts[key]}</span>
                        </button>
                    );
                })}
            </div>

            <div className={styles.toolbar}>
                <div className={styles.filterGroup}>
                    {FILTERS.map((filter) => {
                        const activeValue = searchParamValues[filter.key];
                        const activeLabel = getOptionLabel(filter.options, activeValue);
                        const isOpen = openMenu === filter.key;

                        return (
                            <div key={filter.key} className={styles.menuWrap}>
                                <button
                                    className={`${styles.chip} ${activeValue ? styles.chipActive : ""}`}
                                    type="button"
                                    aria-expanded={isOpen}
                                    onClick={() => setOpenMenu(isOpen ? null : filter.key)}
                                >
                                    {filter.key === "media_type" && (activeValue === "tv" ? <Tv size={15}/> :
                                        <Film size={15}/>)}
                                    {filter.key === "release_status" && <Star size={15} aria-hidden="true"/>}
                                    {filter.key === "year" && <CalendarDays size={15} aria-hidden="true"/>}
                                    {filter.key === "genre_id" && <ListFilter size={15} aria-hidden="true"/>}
                                    <span>{activeValue ? activeLabel : filter.label}</span>
                                    <ChevronDown size={14} aria-hidden="true"/>
                                </button>

                                {isOpen && (
                                    <div className={styles.menu} role="menu">
                                        {filter.options.map((option) => (
                                            <button
                                                key={option.value ?? "all"}
                                                className={`${styles.menuItem} ${option.value === activeValue ? styles.menuItemActive : ""}`}
                                                type="button"
                                                role="menuitemradio"
                                                aria-checked={option.value === activeValue}
                                                onClick={() => updateParam(filter.key, option.value)}
                                            >
                                                <span className={styles.box} aria-hidden="true"/>
                                                <span>{option.label}</span>
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        );
                    })}

                    {activeFilterCount > 0 && (
                        <button className={styles.clearButton} type="button" onClick={clearFilters}>
                            <X size={15} aria-hidden="true"/>
                            Сбросить
                        </button>
                    )}
                </div>

                <div className={`${styles.menuWrap} ${styles.sortWrap}`}>
                    <button
                        className={`${styles.chip} ${styles.sortButton}`}
                        type="button"
                        aria-expanded={openMenu === "sort"}
                        onClick={() => setOpenMenu(openMenu === "sort" ? null : "sort")}
                    >
                        <SortAsc size={16} aria-hidden="true"/>
                        <span>{activeSortLabel}</span>
                        <ChevronDown size={14} aria-hidden="true"/>
                    </button>

                    {openMenu === "sort" && (
                        <div className={`${styles.menu} ${styles.sortMenu}`} role="menu">
                            {SORT_OPTIONS.map((option) => (
                                <button
                                    key={option.value}
                                    className={`${styles.menuItem} ${option.value === selectedSort ? styles.menuItemActive : ""}`}
                                    type="button"
                                    role="menuitemradio"
                                    aria-checked={option.value === selectedSort}
                                    onClick={() => updateParam("sort", option.value)}
                                >
                                    <span className={styles.sortDirection}
                                          aria-hidden="true">{getSortIcon(option.value)}</span>
                                    <span>{option.label}</span>
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            {isError && <p className={styles.stateText} role="alert">Ошибка загрузки списка</p>}

            {!isError && isLoading && (
                <div className={styles.grid} aria-label="Загрузка списка">
                    {Array.from({length: 10}, (_, index) => (
                        <MediaCard key={index} isLoading />
                    ))}
                </div>
            )}

            {!isError && !isLoading && items.length === 0 && (
                <div className={styles.emptyState}>
                    <h2>Список пуст</h2>
                    <p>Добавьте фильмы или сериалы со страницы деталей, чтобы они появились здесь.</p>
                </div>
            )}

            {!isError && !isLoading && items.length > 0 && (
                <div className={styles.grid}>
                    {items.map((media) => (
                        <MediaCard key={`${media.type}-${media.id}`} {...media} type={media.type} withTypeBadge={true}/>
                    ))}
                </div>
            )}

            <Pagination
                currentPage={page}
                totalPages={totalPages}
                pathname="/watchlist"
                searchParams={searchParamValues}
            />
        </section>
    );
};

export default WatchlistClient;
