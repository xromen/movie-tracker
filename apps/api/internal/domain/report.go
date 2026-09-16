package domain

import "time"

type Report struct {
	ID                   int64
	UserID               int64
	PeriodFrom           time.Time
	PeriodTo             time.Time
	CreatedAt            time.Time
	NextScheduleCreateAt time.Time
}

type DueReport struct {
	LatestReportID int64
	UserID         int64
	PeriodFrom     time.Time
	PeriodTo       time.Time
	Interval       int
}

type ReportMovie struct {
	ID        int64
	TMDBID    int64
	Title     string
	Overview  string
	ReleaseAt time.Time
}

type ReportTvEpisode struct {
	ID              int64
	TMDBID          int64
	SeriesTitle     string
	SeriesOverview  string
	SeasonNumber    int
	EpisodeNumber   int
	EpisodeTitle    string
	EpisodeOverview string
	EpisodeAirDate  time.Time
}

type ReportMessage struct {
	ID            int64
	ReportId      int64
	Message       string
	Position      int
	Attempts      int
	NextAttemptAt time.Time
	SentAt        *time.Time
	LastError     *string
}

type DueReportMessage struct {
	ID             int64
	UserTelegramID int64
	Message        string
	Position       int
}
