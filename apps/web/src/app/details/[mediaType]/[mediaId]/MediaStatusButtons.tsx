"use client"

import { useState } from "react"
import { removeWatchStatus, setWatchStatus } from "@/lib/api/media"
import type { MediaType, WatchStatus } from "@/lib/api/types"
import styles from "./MediaStatusButtons.module.css"

export interface MediaStatusButtonsProps {
    mediaId: number
    mediaType: MediaType
    status?: WatchStatus | null
}

const buttons: { className: string; label: string; status: WatchStatus }[] = [
    { className: styles.watched, label: "✓ Просмотрено", status: "watched" },
    { className: styles.wantToWatch, label: "🔖 Хочу посмотреть", status: "want_to_watch" },
    { className: styles.favorite, label: "♥ Избранное", status: "favorite" },
]

const MediaStatusButtons = ({ mediaId, mediaType, status }: MediaStatusButtonsProps) => {
    const [activeStatus, setActiveStatus] = useState<WatchStatus | undefined | null>(status)
    const [isLoading, setIsLoading] = useState(false)
    const [pendingStatus, setPendingStatus] = useState<WatchStatus | null>(null)
    const [error, setError] = useState("")

    const handleStatusClick = async (clickedStatus: WatchStatus) => {
        const oldStatus = activeStatus
        const newStatus = activeStatus === clickedStatus ? null : clickedStatus

        setActiveStatus(newStatus as WatchStatus | undefined)
        setIsLoading(true)
        setPendingStatus(clickedStatus)
        setError("")

        try {
            if (newStatus === null) {
                await removeWatchStatus(mediaId)
            } else {
                await setWatchStatus(mediaId, mediaType, newStatus)
            }
        } catch (error) {
            console.error("Ошибка обновления статуса просмотра:", error)
            setActiveStatus(oldStatus)
            setError("Не удалось обновить статус. Попробуйте еще раз.")
        } finally {
            setIsLoading(false)
            setPendingStatus(null)
        }
    }

    return (
        <div className={styles.wrapper}>
            {buttons.map((button) => (
                <button
                    key={button.status}
                    className={`${button.className} ${activeStatus === button.status ? styles.active : ""}`}
                    onClick={() => handleStatusClick(button.status)}
                    disabled={isLoading}
                    aria-pressed={activeStatus === button.status}
                >
                    {button.label}
                    {isLoading && pendingStatus === button.status && " ..."}
                </button>
            ))}
            {error && <p className={styles.error} role="alert">{error}</p>}
        </div>
    )
}

export default MediaStatusButtons