package domain

import (
	"time"
)

// type TVShow struct {
// 	ID          int64
// 	Title       string
// 	Overview    string
// 	ReleaseDate string
// 	PosterPath  string
// 	Popularity  float32
// 	VoteAverage float32
// }

// type UserTVShow struct {
// 	TVShowID int64
// 	TVShow   *TVShow
// 	Status   WatchStatus
// 	Rating   *int
// }

type TVShowDetail struct {
	TMDBID                 int64 `json:"ID"`
	Title                  string
	Overview               string
	ReleaseDate            string
	LastEpisodeReleaseDate string
	NextEpisodeReleaseDate string
	PosterPath             string
	Popularity             float32
	VoteAverage            float32
	VoteCount              int64
	UserRating             *int
	UserStatus             WatchStatus
	Genres                 []Genre
	OriginalLanguage       string
	OriginCountry          []string
	ProductionCountries    []string
	ProductionCompanies    []ProductionCompany
	OriginalTitle          string
	Status                 string
	Videos                 []Video
	NumberOfSeasons        int
	NumberOfEpisodes       int
	Seasons                []Season
}

type Season struct {
	TMDBID       int64 `json:"ID"`
	ReleaseDate  string
	EpisodeCount int
	Title        string
	Overview     string
	PosterPath   string
	SeasonNumber int
	VoteAverage  float32
	VoteCount    int64
	IsWatched    *bool
}

type Episode struct {
	TMDBID        int64 `json:"ID"`
	AirDate       *time.Time
	EpisodeNumber int
	Title         string
	Overview      string
	Runtime       int
	SeasonNumber  int
	TVShowTMDBID  int64 `json:"TVShowId"`
	StillPath     string
	VoteAverage   float32
	VoteCount     int64
	IsWatched     *bool
}

type TVSchedule struct {
	Status        string
	SeasonNumbers []int
}
