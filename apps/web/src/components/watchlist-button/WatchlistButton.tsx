"use client";

import { useState, useTransition } from "react";
import { removeWatchStatus, setWatchStatus } from "@/lib/api/media";
import type { MediaType, WatchStatus } from "@/lib/api/types";
import styles from "./WatchlistButton.module.css";

interface WatchlistButtonProps {
  mediaId: number;
  currentStatus?: WatchStatus;
  mediaType: MediaType;
}

const STATUS_CONFIG: Record<WatchStatus, { label: string; icon: string }> = {
  watched: { label: "Просмотрено", icon: "✓" },
  want_to_watch: { label: "Хочу посмотреть", icon: "🔖" },
  favorite: { label: "Избранное", icon: "♥" },
};

const WatchlistButton = ({ mediaId, currentStatus, mediaType }: WatchlistButtonProps) => {
  const [selectedStatus, setSelectedStatus] = useState(currentStatus);
  const [isPending, startTransition] = useTransition();

  const handleSelect = (status: WatchStatus) => {
    startTransition(async () => {
      if (selectedStatus === status) {
        try {
          await removeWatchStatus(mediaId);
          setSelectedStatus(undefined);
        } catch {
          return;
        }

        return;
      }

      try {
        await setWatchStatus(mediaId, mediaType, status);
        setSelectedStatus(status);
      } catch {
        return;
      }
    });
  };

  return (
    <div className={styles.container}>
      {(Object.keys(STATUS_CONFIG) as WatchStatus[]).map((status) => (
        <button
          key={status}
          type="button"
          className={`${styles.button} ${selectedStatus === status ? styles.active : ""}`}
          onClick={() => handleSelect(status)}
          disabled={isPending}
          data-tooltip={STATUS_CONFIG[status].label}
        >
          {STATUS_CONFIG[status].icon}
        </button>
      ))}
    </div>
  );
};

export default WatchlistButton;
