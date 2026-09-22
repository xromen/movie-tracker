import type { Metadata } from "next"
import { notFound } from "next/navigation"
import { Building2, ExternalLink, Globe2, MapPin } from "lucide-react"
import MediaCatalog from "@/components/media-catalog/MediaCatalog"
import Pagination from "@/components/pagination/Pagination"
import { getCompanyDetails } from "@/lib/api/company"
import { getMediasListByCompanyId } from "@/lib/api/media"
import type { MediaType } from "@/lib/api/types"
import MediaTypeSelect from "./MediaTypeSelect"
import styles from "./CompanyPage.module.css"

interface CompanyPageProps {
    params: Promise<{
        companyId: string
        mediaType: string
    }>
    searchParams: Promise<{
        page?: string
    }>
}

const isMediaType = (value: string): value is MediaType => value === "movie" || value === "tv"

const getCountryName = (countryCode?: string) => {
    if (!countryCode) {
        return undefined
    }

    try {
        return new Intl.DisplayNames(["ru"], { type: "region" }).of(countryCode) ?? countryCode
    } catch {
        return countryCode
    }
}

const getHomepage = (homepage?: string) => {
    if (!homepage) {
        return undefined
    }

    try {
        const url = new URL(homepage)

        return url.protocol === "http:" || url.protocol === "https:"
            ? { href: url.toString(), label: url.hostname.replace(/^www\./, "") }
            : undefined
    } catch {
        return undefined
    }
}

export const generateMetadata = async ({ params }: CompanyPageProps): Promise<Metadata> => {
    const { companyId } = await params
    const id = Number(companyId)

    if (!Number.isInteger(id) || id <= 0) {
        return { title: "Компания - Movie Tracker" }
    }

    try {
        const company = await getCompanyDetails(id)

        return {
            title: `${company.name} - Movie Tracker`,
            description: company.description || `Фильмы и сериалы компании ${company.name}.`,
        }
    } catch {
        return { title: "Компания - Movie Tracker" }
    }
}

const CompanyPage = async ({ params, searchParams }: CompanyPageProps) => {
    const [{ companyId, mediaType }, query] = await Promise.all([params, searchParams])
    const id = Number(companyId)

    if (!Number.isInteger(id) || id <= 0 || !isMediaType(mediaType)) {
        notFound()
    }

    const page = Math.max(1, Number(query.page ?? 1) || 1)
    const [company, medias] = await Promise.all([
        getCompanyDetails(id),
        getMediasListByCompanyId(mediaType, id, page),
    ])

    const countryName = getCountryName(company.originCountry)
    const homepage = getHomepage(company.homepage)
    const mediaLabel = mediaType === "movie" ? "Фильмы" : "Сериалы"
    const resultsLabel = mediaType === "movie" ? "фильмов" : "сериалов"

    return (
        <div className={styles.page}>
            <section className={styles.hero} aria-labelledby="company-title">
                <div className={styles.logoArea}>
                    {company.logoPath ? (
                        <img className={styles.logo} src={company.logoPath} alt={`Логотип ${company.name}`} />
                    ) : (
                        <Building2 className={styles.logoPlaceholder} size={72} strokeWidth={1.35} aria-hidden="true" />
                    )}
                </div>

                <div className={styles.companyInfo}>
                    <h1 id="company-title" className={styles.title}>{company.name}</h1>

                    {company.description && (
                        <p className={styles.description}>{company.description}</p>
                    )}

                    {(countryName || company.headquarters || homepage) && (
                        <dl className={styles.companyMeta}>
                            {countryName && (
                                <div className={styles.metaItem}>
                                    <Globe2 size={18} aria-hidden="true" />
                                    <div>
                                        <dt>Страна</dt>
                                        <dd>{countryName}</dd>
                                    </div>
                                </div>
                            )}

                            {company.headquarters && (
                                <div className={styles.metaItem}>
                                    <MapPin size={18} aria-hidden="true" />
                                    <div>
                                        <dt>Штаб-квартира</dt>
                                        <dd>{company.headquarters}</dd>
                                    </div>
                                </div>
                            )}

                            {homepage && (
                                <div className={styles.metaItem}>
                                    <ExternalLink size={18} aria-hidden="true" />
                                    <div>
                                        <dt>Официальный сайт</dt>
                                        <dd>
                                            <a href={homepage.href} target="_blank" rel="noreferrer">
                                                {homepage.label}
                                            </a>
                                        </dd>
                                    </div>
                                </div>
                            )}
                        </dl>
                    )}
                </div>
            </section>

            <section className={styles.catalog} aria-labelledby="company-catalog-title">
                <div className={styles.catalogHeader}>
                    <div>
                        <h2 id="company-catalog-title" className={styles.sectionTitle}>{mediaLabel}</h2>
                        <p className={styles.resultsCount}>Всего: {medias.totalItems.toLocaleString("ru-RU")}</p>
                    </div>

                    <MediaTypeSelect companyId={id} mediaType={mediaType} />
                </div>

                {medias.results.length > 0 ? (
                    <MediaCatalog medias={medias.results} />
                ) : (
                    <div className={styles.emptyState}>
                        <Building2 size={32} strokeWidth={1.5} aria-hidden="true" />
                        <p>У этой компании пока нет доступных {resultsLabel}.</p>
                    </div>
                )}

                <Pagination
                    currentPage={page}
                    totalPages={medias.totalPages}
                    pathname={`/company/${id}/${mediaType}`}
                />
            </section>
        </div>
    )
}

export default CompanyPage
