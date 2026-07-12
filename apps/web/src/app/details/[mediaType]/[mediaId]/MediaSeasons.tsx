"use client";

import { useEffect, useState } from "react";
import SeasonList from "@/components/season-list/SeasonList";
import { getTvSeasons } from "@/lib/api/media";
import type { Season } from "@/lib/api/types";

interface MediaSeasonsProps {
  tvId: number;
  isAuthenticated: boolean;
  className?: string;
}

const MediaSeasons = ({ tvId, isAuthenticated, className }: MediaSeasonsProps) => {
  const [seasons, setSeasons] = useState<Season[]>([]);

  useEffect(() => {
    let isMounted = true;

    const loadSeasons = async () => {
      try {
        const loadedSeasons = await getTvSeasons(tvId);

        if (!isMounted) return;

        setSeasons(loadedSeasons);
      } catch {
        if (!isMounted) return;

        return;
      }

    };

    void loadSeasons();

    return () => {
      isMounted = false;
    };
  }, [tvId]);

  return (
    <div className={className}>
      <SeasonList tvId={tvId} seasons={seasons} isAuthenticated={isAuthenticated} />
    </div>
  );
};

export default MediaSeasons;
