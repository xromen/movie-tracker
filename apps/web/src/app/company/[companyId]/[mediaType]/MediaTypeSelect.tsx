"use client"

import Link from "next/link"
import { Check, ChevronDown } from "lucide-react"
import type { FocusEvent } from "react"
import type { MediaType } from "@/lib/api/types"
import styles from "./CompanyPage.module.css"

interface MediaTypeSelectProps {
    companyId: number
    mediaType: MediaType
}

const MediaTypeSelect = ({ companyId, mediaType }: MediaTypeSelectProps) => {
    const options: { label: string; value: MediaType }[] = [
        { label: "Фильмы", value: "movie" },
        { label: "Сериалы", value: "tv" },
    ]
    const selectedOption = options.find((option) => option.value === mediaType) ?? options[0]

    const handleBlur = (event: FocusEvent<HTMLDetailsElement>) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
            event.currentTarget.open = false
        }
    }

    return (
        <div className={styles.mediaTypeControl}>
            <span>Тип контента</span>

            <details className={styles.mediaTypeSelect} onBlur={handleBlur}>
                <summary className={styles.selectTrigger} aria-label={`Тип контента: ${selectedOption.label}`}>
                    <span>{selectedOption.label}</span>
                    <ChevronDown className={styles.selectChevron} size={18} aria-hidden="true" />
                </summary>

                <div className={styles.selectMenu} aria-label="Выбор типа контента">
                    {options.map((option) => {
                        const isSelected = option.value === mediaType

                        return (
                            <Link
                                key={option.value}
                                className={`${styles.selectOption} ${isSelected ? styles.selectOptionActive : ""}`}
                                href={`/company/${companyId}/${option.value}`}
                                aria-current={isSelected ? "page" : undefined}
                            >
                                <span>{option.label}</span>
                                {isSelected && <Check size={16} aria-hidden="true" />}
                            </Link>
                        )
                    })}
                </div>
            </details>
        </div>
    )
}

export default MediaTypeSelect
